package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	appai "github.com/husari/hube/internal/application/ai"
	appapp "github.com/husari/hube/internal/application/app"
	"github.com/husari/hube/internal/application/backup"
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
	apiToken := os.Getenv("HUBE_API_TOKEN")
	if apiToken == "" {
		log.Fatal("HUBE_API_TOKEN is required: the API refuses all requests without it. Set it in your environment or .env file. Generate one with: openssl rand -hex 32")
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

	settingRepo := sqlite.NewSettingRepo(db)
	settingSvc := appsetting.NewService(settingRepo)

	// Seed integration config from env vars (only if not already stored in DB).
	// The DB is the single source of truth at runtime: clients read from it on
	// every request, so changes made through the UI take effect immediately.
	for k, v := range map[string]string{
		"integration.monkeyapi_url":       os.Getenv("MONKEYAPI_URL"),
		"integration.monkeyapi_key":       os.Getenv("MONKEYAPI_KEY"),
		"integration.paypinga_url":        os.Getenv("PAYPINGA_URL"),
		"integration.paypinga_key":        os.Getenv("PAYPINGA_KEY"),
		"integration.claude_api_key":      os.Getenv("ANTHROPIC_API_KEY"),
		"integration.openai_api_key":      os.Getenv("OPENAI_API_KEY"),
		"integration.openai_base_url":     os.Getenv("OPENAI_BASE_URL"),
		"integration.openai_model":        os.Getenv("OPENAI_MODEL"),
		"integration.openai_embedding_model": os.Getenv("OPENAI_EMBEDDING_MODEL"),
	} {
		if err := settingSvc.Seed(ctx, k, v); err != nil {
			log.Printf("warn: seed setting %s: %v", k, err)
		}
	}

	// Log which integrations are configured at boot, based on the seeded DB.
	logBootIntegrations(ctx, settingSvc)

	var ragSvc *appnote.RAGService
	if hasKey, _ := hasIntegrationKey(ctx, settingSvc, "integration.openai_api_key"); hasKey {
		embClient := external.NewEmbeddingsClient(settingSvc)
		ragSvc = appnote.NewRAGService(noteRepo, embClient)
		log.Printf("RAG semantic search enabled")
	}

	// Integration clients always exist; they read their configuration from
	// the settings service on every request.
	moneyMonkey := external.NewMoneyMonkeyClient(settingSvc)
	payPinga := external.NewPayPingaClient(settingSvc)
	claudeClient := external.NewClaudeClient(settingSvc)
	openaiClient := external.NewOpenAIClient(settingSvc)

	var emailSvc *appemail.Service
	if host := os.Getenv("SMTP_HOST"); host != "" {
		smtpClient := external.NewSMTPClient(host, os.Getenv("SMTP_PORT"), os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASS"), os.Getenv("SMTP_FROM"))
		emailSvc = appemail.NewService(smtpClient, taskSvc)
		log.Printf("Email (SMTP) enabled: %s", host)
	}

	hubExecutor := appai.NewHubExecutor(taskSvc, noteSvc, projectSvc, eventSvc, appSvc)

	backupSvc := backup.NewService(db, dbPath, os.Getenv("BACKUP_DIR"), 7)
	go backupSvc.Run(ctx)

	exportSvc := backup.NewExportService(
		sqlite.NewNoteRepo(db),
		sqlite.NewTaskRepo(db),
		sqlite.NewEventRepo(db),
		sqlite.NewAppRepo(db),
		sqlite.NewWishlistRepo(db),
		sqlite.NewProjectRepo(db),
		sqlite.NewDiagramRepo(db),
	)

	origins := []string{"http://localhost:5173", "http://localhost:4173"}
	if domain := os.Getenv("HUBE_DOMAIN"); domain != "" {
		origins = append(origins, "https://"+domain, "http://"+domain)
	}

	router := hubehttp.NewRouter(taskSvc, eventSvc, appSvc, noteSvc, folderSvc, projectSvc, settingSvc, wishlistSvc, diagramSvc, ragSvc, emailSvc, moneyMonkey, payPinga, claudeClient, openaiClient, hubExecutor, backupSvc, exportSvc, origins, apiToken)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("hube API running on %s", addr)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		// Bumped to 5 minutes to accommodate long agentic AI chat streams
		// (multiple tool rounds + streaming text). The /ai/chat handler streams
		// via Flush and context cancellation still cuts off runaway requests.
		WriteTimeout: 5 * time.Minute,
		IdleTimeout:  120 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func hasIntegrationKey(ctx context.Context, svc *appsetting.Service, key string) (bool, string) {
	v, err := svc.Get(ctx, key)
	if err != nil {
		return false, ""
	}
	return v != "", v
}

func logBootIntegrations(ctx context.Context, svc *appsetting.Service) {
	if ok, v := hasIntegrationKey(ctx, svc, "integration.monkeyapi_url"); ok {
		log.Printf("Money Monkey integration enabled: %s", v)
	}
	if ok, v := hasIntegrationKey(ctx, svc, "integration.paypinga_url"); ok {
		log.Printf("PayPinga integration enabled: %s", v)
	}
	if hasIntegrationKeyBool(ctx, svc, "integration.claude_api_key") {
		log.Printf("Anthropic Claude enabled (claude-sonnet-4-6)")
	}
	if hasIntegrationKeyBool(ctx, svc, "integration.openai_api_key") {
		base, _ := svc.Get(ctx, "integration.openai_base_url")
		log.Printf("OpenAI-compatible provider enabled (base: %s)", base)
	}
}

func hasIntegrationKeyBool(ctx context.Context, svc *appsetting.Service, key string) bool {
	ok, _ := hasIntegrationKey(ctx, svc, key)
	return ok
}
