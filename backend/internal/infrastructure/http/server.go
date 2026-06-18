package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	appai "github.com/husari/hube/internal/application/ai"
	appapp "github.com/husari/hube/internal/application/app"
	appevent "github.com/husari/hube/internal/application/event"
	appfolder "github.com/husari/hube/internal/application/folder"
	appnote "github.com/husari/hube/internal/application/note"
	appproject "github.com/husari/hube/internal/application/project"
	appsetting "github.com/husari/hube/internal/application/setting"
	apptask "github.com/husari/hube/internal/application/task"
	"github.com/husari/hube/internal/infrastructure/external"
	"github.com/husari/hube/internal/infrastructure/http/handler"
)

func NewRouter(
	taskSvc *apptask.Service,
	eventSvc *appevent.Service,
	appSvc *appapp.Service,
	noteSvc *appnote.Service,
	folderSvc *appfolder.Service,
	projectSvc *appproject.Service,
	settingSvc *appsetting.Service,
	moneyMonkey *external.MoneyMonkeyClient,
	payPinga *external.PayPingaClient,
	claude  *external.ClaudeClient,
	openai  *external.OpenAIClient,
	hubExec *appai.HubExecutor,
) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173", "http://localhost:4173"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
	}))

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/tasks", handler.NewTaskHandler(taskSvc).Routes())
		r.Route("/events", handler.NewEventHandler(eventSvc).Routes())
		r.Route("/apps", handler.NewAppHandler(appSvc).Routes())
		r.Route("/notes", handler.NewNoteHandler(noteSvc).Routes())
		r.Route("/folders", handler.NewFolderHandler(folderSvc).Routes())
		r.Route("/projects", handler.NewProjectHandler(projectSvc).Routes())
		sh := handler.NewSettingHandler(settingSvc)
		r.Get("/settings", sh.Get)
		r.Put("/settings", sh.Put)
		r.Get("/finance/summary", handler.NewFinanceHandler(moneyMonkey).Summary)
		r.Get("/cards/summary", handler.NewCardTrackerHandler(payPinga).Summary)
		r.Post("/ai/chat", handler.NewAIHandler(claude, openai, hubExec).Chat)
	})

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return r
}
