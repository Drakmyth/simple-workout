package app

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"io"
)

const (
	imageMale        = "male"
	imageFemale      = "female"
	imageRandom      = "random"
	imageAlternating = "alternating"
)

type Exercise struct {
	ID     int64
	Slug   string
	Name   string
	Notes  string
	Steps  []string
	Images map[string]ExerciseImage
}

type ExerciseImage struct {
	Path    string
	AltText string
}

func (a *App) exercises() ([]Exercise, error) {
	rows, err := a.db.Query(`SELECT e.id, e.slug, e.name, e.notes, i.variant, i.path, i.alt_text
		FROM exercises e JOIN exercise_images i ON i.exercise_id = e.id
		ORDER BY e.name, i.variant`)
	if err != nil {
		return nil, fmt.Errorf("query exercises: %w", err)
	}
	defer rows.Close()

	var exercises []Exercise
	var current *Exercise
	for rows.Next() {
		var id int64
		var slug, name, notes, variant string
		var image ExerciseImage
		if err := rows.Scan(&id, &slug, &name, &notes, &variant, &image.Path, &image.AltText); err != nil {
			return nil, fmt.Errorf("scan exercise: %w", err)
		}
		if current == nil || current.Slug != slug {
			exercises = append(exercises, Exercise{ID: id, Slug: slug, Name: name, Notes: notes, Images: make(map[string]ExerciseImage, 2)})
			current = &exercises[len(exercises)-1]
		}
		current.Images[variant] = image
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read exercises: %w", err)
	}
	return exercises, nil
}

func (a *App) exercise(slug string) (Exercise, error) {
	var exercise Exercise
	var id int64
	err := a.db.QueryRow("SELECT id, slug, name, notes FROM exercises WHERE slug = ?", slug).
		Scan(&id, &exercise.Slug, &exercise.Name, &exercise.Notes)
	exercise.ID = id
	if err != nil {
		return Exercise{}, fmt.Errorf("query exercise: %w", err)
	}
	exercise.Images = make(map[string]ExerciseImage, 2)

	rows, err := a.db.Query("SELECT variant, path, alt_text FROM exercise_images WHERE exercise_id = ? ORDER BY variant", id)
	if err != nil {
		return Exercise{}, fmt.Errorf("query exercise images: %w", err)
	}
	for rows.Next() {
		var variant string
		var image ExerciseImage
		if err := rows.Scan(&variant, &image.Path, &image.AltText); err != nil {
			rows.Close()
			return Exercise{}, fmt.Errorf("scan exercise image: %w", err)
		}
		exercise.Images[variant] = image
	}
	if err := rows.Close(); err != nil {
		return Exercise{}, fmt.Errorf("close exercise images: %w", err)
	}
	if err := rows.Err(); err != nil {
		return Exercise{}, fmt.Errorf("read exercise images: %w", err)
	}

	stepRows, err := a.db.Query("SELECT instruction FROM exercise_steps WHERE exercise_id = ? ORDER BY position", id)
	if err != nil {
		return Exercise{}, fmt.Errorf("query exercise steps: %w", err)
	}
	defer stepRows.Close()
	for stepRows.Next() {
		var step string
		if err := stepRows.Scan(&step); err != nil {
			return Exercise{}, fmt.Errorf("scan exercise step: %w", err)
		}
		exercise.Steps = append(exercise.Steps, step)
	}
	if err := stepRows.Err(); err != nil {
		return Exercise{}, fmt.Errorf("read exercise steps: %w", err)
	}
	return exercise, nil
}

func (a *App) imagePreference() (string, error) {
	var preference string
	if err := a.db.QueryRow("SELECT image_preference FROM settings WHERE id = 1").Scan(&preference); err != nil {
		return "", fmt.Errorf("query image preference: %w", err)
	}
	return preference, nil
}

func (a *App) setImagePreference(preference string) error {
	if !validImagePreference(preference) {
		return fmt.Errorf("invalid image preference %q", preference)
	}
	if _, err := a.db.Exec("UPDATE settings SET image_preference = ? WHERE id = 1", preference); err != nil {
		return fmt.Errorf("update image preference: %w", err)
	}
	return nil
}

func validImagePreference(preference string) bool {
	switch preference {
	case imageMale, imageFemale, imageRandom, imageAlternating:
		return true
	default:
		return false
	}
}

func selectImageVariant(preference, alternatingNext string) (variant, next string, err error) {
	switch preference {
	case imageMale, imageFemale:
		return preference, alternatingNext, nil
	case imageAlternating:
		if alternatingNext != imageFemale {
			return imageMale, imageFemale, nil
		}
		return imageFemale, imageMale, nil
	case imageRandom:
		return randomImageVariant(rand.Reader)
	default:
		return "", alternatingNext, fmt.Errorf("invalid image preference %q", preference)
	}
}

func randomImageVariant(source io.Reader) (variant, next string, err error) {
	var value [1]byte
	if _, err := io.ReadFull(source, value[:]); err != nil {
		return "", "", fmt.Errorf("choose random image variant: %w", err)
	}
	if value[0]&1 == 0 {
		return imageMale, "", nil
	}
	return imageFemale, "", nil
}

func isNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
