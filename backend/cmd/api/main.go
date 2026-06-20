package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"time"

	appai "github.com/husari/hube/internal/application/ai"
	appapp "github.com/husari/hube/internal/application/app"
	appdiagram "github.com/husari/hube/internal/application/diagram"
	appemail "github.com/husari/hube/internal/application/email"
	appevent "github.com/husari/hube/internal/application/event"
	appfolder "github.com/husari/hube/internal/application/folder"
	appnote "github.com/husari/hube/internal/application/note"
	appproject "github.com/husari/hube/internal/application/project"
	appsetting "github.com/husari/hube/internal/application/setting"
	apptask "github.com/husari/hube/internal/application/task"
	appwishlist "github.com/husari/hube/internal/application/wishlist"
	"github.com/husari/hube/internal/infrastructure/external"
	hubehttp "github.com/husari/hube/internal/infrastructure/http"
	"github.com/husari/hube/internal/infrastructure/sqlite"
)

func main() {
	if err := godotenv.Load(); err == nil {
		log.Println("loaded .env")
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./hube.db"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db, err := sqlite.Open(dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	taskRepo := sqlite.NewTaskRepo(db)
	taskSvc := apptask.NewService(taskRepo)
	go apptask.NewScheduler(taskRepo).Run(ctx)
	eventSvc := appevent.NewService(sqlite.NewEventRepo(db))
	appSvc := appapp.NewService(sqlite.NewAppRepo(db))
	noteRepo := sqlite.NewNoteRepo(db)
	noteSvc := appnote.NewService(noteRepo)
	folderSvc := appfolder.NewService(sqlite.NewFolderRepo(db))
	projectSvc := appproject.NewService(sqlite.NewProjectRepo(db))
	wishlistSvc := appwishlist.NewService(sqlite.NewWishlistRepo(db))
	diagramSvc := appdiagram.NewService(sqlite.NewDiagramRepo(db))

	var ragSvc *appnote.RAGService
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		embClient := external.NewEmbeddingsClient(key, os.Getenv("OPENAI_BASE_URL"), os.Getenv("OPENAI_EMBEDDING_MODEL"))
		ragSvc = appnote.NewRAGService(noteRepo, embClient)
		log.Printf("RAG semantic search enabled")
	}

	settingRepo := sqlite.NewSettingRepo(db)
	settingSvc := appsetting.NewService(settingRepo)

	// Seed integration config from env vars (only if not already stored in DB)
	for k, v := range map[string]string{
		"integration.monkeyapi_url":  os.Getenv("MONKEYAPI_URL"),
		"integration.monkeyapi_key":  os.Getenv("MONKEYAPI_KEY"),
		"integration.paypinga_url":   os.Getenv("PAYPINGA_URL"),
		"integration.paypinga_key":   os.Getenv("PAYPINGA_KEY"),
		"integration.claude_api_key":  os.Getenv("ANTHROPIC_API_KEY"),
		"integration.openai_api_key":  os.Getenv("OPENAI_API_KEY"),
		"integration.openai_base_url": os.Getenv("OPENAI_BASE_URL"),
		"integration.openai_model":    os.Getenv("OPENAI_MODEL"),
	} {
		if err := settingSvc.Seed(ctx, k, v); err != nil {
			log.Printf("warn: seed setting %s: %v", k, err)
		}
	}

	var moneyMonkey *external.MoneyMonkeyClient
	if url, key := os.Getenv("MONKEYAPI_URL"), os.Getenv("MONKEYAPI_KEY"); url != "" && key != "" {
		moneyMonkey = external.NewMoneyMonkeyClient(url, key)
		log.Printf("Money Monkey integration enabled: %s", url)
	}

	var payPinga *external.PayPingaClient
	if url, key := os.Getenv("PAYPINGA_URL"), os.Getenv("PAYPINGA_KEY"); url != "" && key != "" {
		payPinga = external.NewPayPingaClient(url, key)
		log.Printf("PayPinga integration enabled: %s", url)
	}

	var claudeClient *external.ClaudeClient
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		claudeClient = external.NewClaudeClient(key)
		log.Printf("Anthropic Claude enabled (claude-sonnet-4-6)")
	}

	var openaiClient *external.OpenAIClient
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		openaiClient = external.NewOpenAIClient(key, os.Getenv("OPENAI_BASE_URL"), os.Getenv("OPENAI_MODEL"))
		log.Printf("OpenAI-compatible provider enabled (base: %s)", os.Getenv("OPENAI_BASE_URL"))
	}

	var emailSvc *appemail.Service
	if host := os.Getenv("SMTP_HOST"); host != "" {
		smtpClient := external.NewSMTPClient(host, os.Getenv("SMTP_PORT"), os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASS"), os.Getenv("SMTP_FROM"))
		emailSvc = appemail.NewService(smtpClient, taskSvc)
		log.Printf("Email (SMTP) enabled: %s", host)
	}

	hubExecutor := appai.NewHubExecutor(taskSvc, noteSvc, projectSvc, eventSvc, appSvc)

	origins := []string{"http://localhost:5173", "http://localhost:4173"}
	if domain := os.Getenv("HUBE_DOMAIN"); domain != "" {
		origins = append(origins, "https://"+domain, "http://"+domain)
	}

	router := hubehttp.NewRouter(taskSvc, eventSvc, appSvc, noteSvc, folderSvc, projectSvc, settingSvc, wishlistSvc, diagramSvc, ragSvc, emailSvc, moneyMonkey, payPinga, claudeClient, openaiClient, hubExecutor, origins)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("hube API running on %s", addr)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
