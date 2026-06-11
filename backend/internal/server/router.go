// Package server wires the domain handlers into an http.Handler. Keeping this
// separate from main lets tests spin up the full router against a test DB.
package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tech4mation/tasks-api/internal/auth"
	"github.com/tech4mation/tasks-api/internal/config"
	"github.com/tech4mation/tasks-api/internal/events"
	"github.com/tech4mation/tasks-api/internal/httpx"
	"github.com/tech4mation/tasks-api/internal/middleware"
	"github.com/tech4mation/tasks-api/internal/task"
	"github.com/tech4mation/tasks-api/internal/user"
)

// New builds the application router with all routes mounted.
func New(pool *pgxpool.Pool, cfg *config.Config) http.Handler {
	issuer := auth.NewIssuer(cfg.JWTSecret, cfg.JWTTTL)
	broker := events.NewBroker()

	userRepo := user.NewRepo(pool)
	taskRepo := task.NewRepo(pool)

	authHandler := user.NewHandler(userRepo, issuer, cfg)
	taskHandler := task.NewHandler(taskRepo, broker)

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.FrontendOrigin},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/auth", func(r chi.Router) {
		r.Post("/signup", authHandler.Signup)
		r.Post("/login", authHandler.Login)
		r.Post("/logout", authHandler.Logout)
		r.With(middleware.Authenticate(issuer)).Get("/me", authHandler.Me)
	})

	r.Route("/tasks", func(r chi.Router) {
		r.Use(middleware.Authenticate(issuer))
		r.Post("/", taskHandler.Create)
		r.Get("/", taskHandler.List)
		r.Get("/stream", taskHandler.Stream)
		r.Get("/{id}", taskHandler.Get)
		r.Patch("/{id}", taskHandler.Update)
		r.Delete("/{id}", taskHandler.Delete)
	})

	r.Route("/admin", func(r chi.Router) {
		r.Use(middleware.Authenticate(issuer))
		r.Use(middleware.RequireAdmin)
		r.Get("/tasks", taskHandler.ListAll)
	})

	return r
}
