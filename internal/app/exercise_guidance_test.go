package app

import (
	"context"
	"database/sql"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

func assertExerciseGuidance(t *testing.T, store *Store) {
	t.Helper()
	deadBug, err := store.exercise("dead-bug")
	if err != nil {
		t.Fatal(err)
	}
	if deadBug.Steps[1] != "Engage your core, then extend one arm and the opposite leg while keeping your lower back gently supported by the floor." {
		t.Fatalf("dead bug step = %q", deadBug.Steps[1])
	}
	if deadBug.Notes != "Avoid arching your lower back. Only extend your arm and leg as far as you can while maintaining core engagement." {
		t.Fatalf("dead bug notes = %q", deadBug.Notes)
	}
	press, err := store.exercise("dumbbell-floor-press")
	if err != nil {
		t.Fatal(err)
	}
	if press.Steps[2] != "Press the dumbbells over your chest until your arms are straight, then lower under control until your upper arms touch down gently." {
		t.Fatalf("floor press step = %q", press.Steps[2])
	}
	if press.Notes != "Keep your shoulder blades supported on the floor rather than reaching them toward the ceiling.\nPause gently when your upper arms reach the floor; do not bounce your elbows." {
		t.Fatalf("floor press notes = %q", press.Notes)
	}
}

func TestExerciseGuidanceFreshDatabase(t *testing.T) {
	assertExerciseGuidance(t, newTestApp(t).store)
}

func TestExerciseGuidanceUpgrade(t *testing.T) {
	dir := t.TempDir()
	db, err := sql.Open("sqlite", filepath.Join(dir, "simple-workout.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE schema_migrations (name TEXT PRIMARY KEY, applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP) STRICT`); err != nil {
		t.Fatal(err)
	}
	paths, err := fs.Glob(migrations, "migrations/*.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		if strings.HasSuffix(path, "007_exercise_guidance.sql") {
			break
		}
		data, err := migrations.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(string(data)); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec("INSERT INTO schema_migrations (name) VALUES (?)", filepath.Base(path)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec("UPDATE workouts SET name = 'My workout' WHERE id = 1; UPDATE settings SET rest_seconds = 17"); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	store, err := newStore(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	assertExerciseGuidance(t, store)
	if err := store.applyMigrations(context.Background()); err != nil {
		t.Fatal(err)
	}
	assertExerciseGuidance(t, store)
	workout, err := store.playbackWorkout(1)
	if err != nil || workout.Name != "My workout" {
		t.Fatalf("workout = %+v, %v", workout, err)
	}
	settings, err := store.playbackSettings()
	if err != nil || settings.RestSeconds != 17 {
		t.Fatalf("settings = %+v, %v", settings, err)
	}
}
