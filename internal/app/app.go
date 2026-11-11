package app

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
)

//go:embed migrations/*.sql templates/*.html static/*
var files embed.FS

type Config struct {
	DataDir string
	Logger  *slog.Logger
}

type App struct {
	store     *Store
	handler   http.Handler
	templates *template.Template
	logger    *slog.Logger
}

func New(ctx context.Context, config Config) (*App, error) {
	if config.Logger == nil {
		config.Logger = slog.Default()
	}
	if err := os.MkdirAll(config.DataDir, 0o750); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}

	store, err := newStore(ctx, config.DataDir)
	if err != nil {
		return nil, err
	}

	templates, err := template.ParseFS(files, "templates/*.html")
	if err != nil {
		store.Close()
		return nil, fmt.Errorf("parse templates: %w", err)
	}
	static, err := fs.Sub(files, "static")
	if err != nil {
		store.Close()
		return nil, fmt.Errorf("open static files: %w", err)
	}

	application := &App{store: store, templates: templates, logger: config.Logger}
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(static)))
	mux.HandleFunc("GET /service-worker.js", serviceWorker)
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("GET /exercises", application.exerciseList)
	mux.HandleFunc("GET /exercises/{slug}", application.exerciseDetail)
	mux.HandleFunc("POST /exercises/image-variant", application.updateGuideVariant)
	mux.HandleFunc("GET /workouts", application.workoutList)
	mux.HandleFunc("GET /workouts/{id}/edit", application.workoutEdit)
	mux.HandleFunc("POST /workouts/{id}/rename", application.workoutRename)
	mux.HandleFunc("POST /workouts/{id}/exercises", application.workoutExerciseAdd)
	mux.HandleFunc("POST /workouts/{id}/exercises/{position}/update", application.workoutExerciseUpdate)
	mux.HandleFunc("POST /workouts/{id}/exercises/{position}/remove", application.workoutExerciseRemove)
	mux.HandleFunc("POST /workouts/{id}/exercises/{position}/move", application.workoutExerciseMove)
	mux.HandleFunc("GET /workouts/{id}/start", application.playbackSetup)
	mux.HandleFunc("POST /workouts/{id}/play", application.playbackStart)
	mux.HandleFunc("POST /workouts/{id}/complete", application.playbackComplete)
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := templates.ExecuteTemplate(w, "home.html", nil); err != nil {
			config.Logger.Error("render home page", "error", err)
		}
	})

	application.handler = securityHeaders(mux)
	return application, nil
}

func (a *App) Handler() http.Handler {
	return a.handler
}

func (a *App) Close() error {
	return a.store.Close()
}

func serviceWorker(w http.ResponseWriter, _ *http.Request) {
	worker, err := files.ReadFile("static/service-worker.js")
	if err != nil {
		http.Error(w, "Service worker unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write(worker)
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok\n"))
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self'; script-src 'self'")
		next.ServeHTTP(w, r)
	})
}
