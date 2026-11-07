package app

import (
	"bytes"
	"database/sql"
	"testing"
)

func TestSeededExerciseCatalog(t *testing.T) {
	application := newTestApp(t)
	exercises, err := application.exercises()
	if err != nil {
		t.Fatalf("exercises() error = %v", err)
	}
	if len(exercises) != 9 {
		t.Fatalf("exercise count = %d, want 9", len(exercises))
	}
	for _, exercise := range exercises {
		if len(exercise.Images) != 2 {
			t.Errorf("%s image count = %d, want 2", exercise.Slug, len(exercise.Images))
		}
		if exercise.Images[imageMale].Path == "" || exercise.Images[imageFemale].Path == "" {
			t.Errorf("%s is missing a model variant", exercise.Slug)
		}
	}

	deadBug, err := application.exercise("dead-bug")
	if err != nil {
		t.Fatalf("exercise(dead-bug) error = %v", err)
	}
	if deadBug.Name != "Dead bug" || len(deadBug.Steps) != 3 {
		t.Fatalf("dead bug = (%q, %d steps), want (Dead bug, 3 steps)", deadBug.Name, len(deadBug.Steps))
	}
	if _, err := application.exercise("missing"); !isNotFound(err) {
		t.Fatalf("exercise(missing) error = %v, want sql.ErrNoRows", err)
	}
}

func TestImagePreference(t *testing.T) {
	application := newTestApp(t)
	preference, err := application.imagePreference()
	if err != nil {
		t.Fatalf("imagePreference() error = %v", err)
	}
	if preference != imageAlternating {
		t.Fatalf("initial preference = %q, want %q", preference, imageAlternating)
	}
	if err := application.setImagePreference(imageFemale); err != nil {
		t.Fatalf("setImagePreference() error = %v", err)
	}
	preference, err = application.imagePreference()
	if err != nil || preference != imageFemale {
		t.Fatalf("updated preference = (%q, %v), want (%q, nil)", preference, err, imageFemale)
	}
	if err := application.setImagePreference("unsupported"); err == nil {
		t.Fatal("setImagePreference(unsupported) error = nil")
	}
}

func TestSelectImageVariant(t *testing.T) {
	tests := []struct {
		name       string
		preference string
		next       string
		want       string
		wantNext   string
	}{
		{name: "male", preference: imageMale, next: imageFemale, want: imageMale, wantNext: imageFemale},
		{name: "female", preference: imageFemale, next: imageMale, want: imageFemale, wantNext: imageMale},
		{name: "alternating starts male", preference: imageAlternating, want: imageMale, wantNext: imageFemale},
		{name: "alternating continues female", preference: imageAlternating, next: imageFemale, want: imageFemale, wantNext: imageMale},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, next, err := selectImageVariant(test.preference, test.next)
			if err != nil || got != test.want || next != test.wantNext {
				t.Fatalf("selectImageVariant() = (%q, %q, %v), want (%q, %q, nil)", got, next, err, test.want, test.wantNext)
			}
		})
	}
	if _, _, err := selectImageVariant("unsupported", ""); err == nil {
		t.Fatal("selectImageVariant(unsupported) error = nil")
	}
}

func TestRandomImageVariant(t *testing.T) {
	male, _, err := randomImageVariant(bytes.NewReader([]byte{2}))
	if err != nil || male != imageMale {
		t.Fatalf("even random variant = (%q, %v), want (%q, nil)", male, err, imageMale)
	}
	female, _, err := randomImageVariant(bytes.NewReader([]byte{3}))
	if err != nil || female != imageFemale {
		t.Fatalf("odd random variant = (%q, %v), want (%q, nil)", female, err, imageFemale)
	}
	if _, _, err := randomImageVariant(bytes.NewReader(nil)); err == nil {
		t.Fatal("empty random source error = nil")
	}
}

func TestMissingExerciseUsesSQLNotFound(t *testing.T) {
	if !isNotFound(sql.ErrNoRows) {
		t.Fatal("isNotFound(sql.ErrNoRows) = false")
	}
}
