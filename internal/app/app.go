package app

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql templates/*.html static/*
var files embed.FS

type Config struct {
	DataDir string
	Logger  *slog.Logger
}

type App struct {
	db      *sql.DB
	handler http.Handler
}

func New(ctx context.Context, config Config) (*App, error) {
	if config.Logger == nil {
		config.Logger = slog.Default()
	}
	if err := os.MkdirAll(config.DataDir, 0o750); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(config.DataDir, "simple-workout.db"))
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(1)

	if err := initializeDatabase(ctx, db); err != nil {
		db.Close()
		return nil, err
	}

	templates, err := template.ParseFS(files, "templates/*.html")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("parse templates: %w", err)
	}
	static, err := fs.Sub(files, "static")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("open static files: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(static)))
	mux.HandleFunc("GET /health", health)
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

	return &App{db: db, handler: securityHeaders(mux)}, nil
}

func (a *App) Handler() http.Handler {
	return a.handler
}

func (a *App) Close() error {
	return a.db.Close()
}

func initializeDatabase(ctx context.Context, db *sql.DB) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if _, err := db.ExecContext(ctx, "PRAGMA journal_mode = WAL; PRAGMA foreign_keys = ON;"); err != nil {
		return fmt.Errorf("configure database: %w", err)
	}
	return applyMigrations(ctx, db)
}

func applyMigrations(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		name TEXT PRIMARY KEY,
		applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	) STRICT`); err != nil {
		return fmt.Errorf("create migration ledger: %w", err)
	}

	rows, err := db.QueryContext(ctx, "SELECT name FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("query applied migrations: %w", err)
	}
	applied := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			return fmt.Errorf("scan applied migration: %w", err)
		}
		applied[name] = true
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("close applied migrations: %w", err)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read applied migrations: %w", err)
	}

	paths, err := fs.Glob(files, "migrations/*.sql")
	if err != nil {
		return fmt.Errorf("find migrations: %w", err)
	}
	slices.Sort(paths)
	for _, path := range paths {
		name := filepath.Base(path)
		if applied[name] {
			continue
		}
		migration, err := files.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx, string(migration)); err != nil {
			tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (name) VALUES (?)", name); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %s: %w", name, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}
	}
	return nil
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
