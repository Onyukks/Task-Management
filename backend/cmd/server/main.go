// Command server boots the task-management REST API.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tech4mation/tasks-api/internal/auth"
	"github.com/tech4mation/tasks-api/internal/config"
	"github.com/tech4mation/tasks-api/internal/db"
	"github.com/tech4mation/tasks-api/internal/server"
	"github.com/tech4mation/tasks-api/internal/user"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuration error", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := db.Migrate(pool); err != nil {
		slog.Error("migrations failed", "error", err)
		os.Exit(1)
	}
	slog.Info("migrations applied")

	seedDemoAdmin(ctx, pool, cfg)

	handler := server.New(pool, cfg)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           withRequestLogging(handler),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("server listening", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
	}
}

// seedDemoAdmin creates a demo admin account when SEED_ADMIN_* are configured,
// so reviewers can try the admin "view all tasks" feature without running SQL.
func seedDemoAdmin(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config) {
	if cfg.SeedAdminEmail == "" || cfg.SeedAdminPassword == "" {
		return
	}
	hash, err := auth.HashPassword(cfg.SeedAdminPassword)
	if err != nil {
		slog.Error("seed admin: hash failed", "error", err)
		return
	}
	if err := user.NewRepo(pool).EnsureAdmin(ctx, cfg.SeedAdminEmail, "Demo Admin", hash); err != nil {
		slog.Error("seed admin failed", "error", err)
		return
	}
	slog.Info("demo admin ensured", "email", cfg.SeedAdminEmail)
}

// withRequestLogging logs each request with method, path, status, and duration.
func withRequestLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}
