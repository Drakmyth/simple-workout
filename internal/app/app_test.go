package app

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestApp(t *testing.T) *App {
	t.Helper()
	application, err := New(context.Background(), Config{
		DataDir: t.TempDir(),
		Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	t.Cleanup(func() { _ = application.Close() })
	return application
}

func TestHealth(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()
	newTestApp(t).Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK || response.Body.String() != "ok\n" {
		t.Fatalf("health response = (%d, %q)", response.Code, response.Body.String())
	}
}

func TestHome(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()
	newTestApp(t).Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if !strings.Contains(response.Body.String(), "Your workouts, one set at a time.") {
		t.Fatal("home page did not contain heading")
	}
}

func TestInitialSettings(t *testing.T) {
	application := newTestApp(t)
	var restSeconds int
	var weightUnit string
	if err := application.db.QueryRow("SELECT rest_seconds, weight_unit FROM settings WHERE id = 1").Scan(&restSeconds, &weightUnit); err != nil {
		t.Fatalf("query settings: %v", err)
	}
	if restSeconds != 60 || weightUnit != "kg" {
		t.Fatalf("settings = (%d, %q), want (60, kg)", restSeconds, weightUnit)
	}
}

func TestMigrationsAreRecordedAndIdempotent(t *testing.T) {
	application := newTestApp(t)
	assertMigrationCount(t, application, 3)

	if err := applyMigrations(context.Background(), application.db); err != nil {
		t.Fatalf("applyMigrations() error = %v", err)
	}
	assertMigrationCount(t, application, 3)
}

func assertMigrationCount(t *testing.T, application *App, want int) {
	t.Helper()
	var count int
	if err := application.db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	if count != want {
		t.Fatalf("migration count = %d, want %d", count, want)
	}
}
