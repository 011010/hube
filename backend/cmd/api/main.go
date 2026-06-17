package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	appapp "github.com/husari/hube/internal/application/app"
	appevent "github.com/husari/hube/internal/application/event"
	apptask "github.com/husari/hube/internal/application/task"
	hubehttp "github.com/husari/hube/internal/infrastructure/http"
	"github.com/husari/hube/internal/infrastructure/external"
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

	var moneyMonkey *external.MoneyMonkeyClient
	if url, key := os.Getenv("MONKEYAPI_URL"), os.Getenv("MONKEYAPI_KEY"); url != "" && key != "" {
		moneyMonkey = external.NewMoneyMonkeyClient(url, key)
		log.Printf("Money Monkey integration enabled: %s", url)
	}

	router := hubehttp.NewRouter(taskSvc, eventSvc, appSvc, moneyMonkey)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("hube API running on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}
