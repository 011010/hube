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

	router := hubehttp.NewRouter(taskSvc, eventSvc, appSvc)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("hube API running on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}
