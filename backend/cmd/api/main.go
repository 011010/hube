package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	appai "github.com/husari/hube/internal/application/ai"
	appapp "github.com/husari/hube/internal/application/app"
	appevent "github.com/husari/hube/internal/application/event"
	appfolder "github.com/husari/hube/internal/application/folder"
	appnote "github.com/husari/hube/internal/application/note"
	appproject "github.com/husari/hube/internal/application/project"
	appsetting "github.com/husari/hube/internal/application/setting"
	apptask "github.com/husari/hube/internal/application/task"
	"github.com/husari/hube/internal/infrastructure/external"
	hubehttp "github.com/husari/hube/internal/infrastructure/http"
	"github.com/husari/hube/internal/infrastructure/sqlite"
)

func main() {
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

	taskSvc := apptask.NewService(sqlite.NewTaskRepo(db))
	eventSvc := appevent.NewService(sqlite.NewEventRepo(db))
	appSvc := appapp.NewService(sqlite.NewAppRepo(db))
	noteSvc := appnote.NewService(sqlite.NewNoteRepo(db))
	folderSvc := appfolder.NewService(sqlite.NewFolderRepo(db))
	projectSvc := appproject.NewService(sqlite.NewProjectRepo(db))

	settingRepo := sqlite.NewSettingRepo(db)
	settingSvc := appsetting.NewService(settingRepo)

	// Seed integration config from env vars (only if not already stored in DB)
	ctx := context.Background()
	for k, v := range map[string]string{
		"integration.monkeyapi_url":  os.Getenv("MONKEYAPI_URL"),
		"integration.monkeyapi_key":  os.Getenv("MONKEYAPI_KEY"),
		"integration.paypinga_url":   os.Getenv("PAYPINGA_URL"),
		"integration.paypinga_key":   os.Getenv("PAYPINGA_KEY"),
		"integration.claude_api_key": os.Getenv("ANTHROPIC_API_KEY"),
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

	claudeKey := os.Getenv("ANTHROPIC_API_KEY")
	var claudeClient *external.ClaudeClient
	if claudeKey != "" {
		claudeClient = external.NewClaudeClient(claudeKey)
		log.Printf("Claude AI integration enabled (model: claude-sonnet-4-6)")
	}
	hubExecutor := appai.NewHubExecutor(taskSvc, noteSvc, projectSvc, eventSvc, appSvc)

	router := hubehttp.NewRouter(taskSvc, eventSvc, appSvc, noteSvc, folderSvc, projectSvc, settingSvc, moneyMonkey, payPinga, claudeClient, hubExecutor)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("hube API running on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}
