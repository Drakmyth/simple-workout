package app

import "fmt"

func (s *Store) exercises() ([]Exercise, error) {
	rows, err := s.db.Query(`SELECT e.id, e.slug, e.name, e.notes, i.variant, i.path, i.alt_text
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

func (s *Store) exercise(slug string) (Exercise, error) {
	var exercise Exercise
	var id int64
	err := s.db.QueryRow("SELECT id, slug, name, notes FROM exercises WHERE slug = ?", slug).
		Scan(&id, &exercise.Slug, &exercise.Name, &exercise.Notes)
	exercise.ID = id
	if err != nil {
		return Exercise{}, fmt.Errorf("query exercise: %w", err)
	}
	exercise.Images, err = s.exerciseImages(id)
	if err != nil {
		return Exercise{}, err
	}
	exercise.Steps, err = s.exerciseSteps(id)
	if err != nil {
		return Exercise{}, err
	}
	return exercise, nil
}

func (s *Store) exerciseImages(id int64) (map[string]ExerciseImage, error) {
	rows, err := s.db.Query("SELECT variant, path, alt_text FROM exercise_images WHERE exercise_id = ? ORDER BY variant", id)
	if err != nil {
		return nil, fmt.Errorf("query exercise images: %w", err)
	}
	defer rows.Close()
	images := make(map[string]ExerciseImage, 2)
	for rows.Next() {
		var variant string
		var image ExerciseImage
		if err := rows.Scan(&variant, &image.Path, &image.AltText); err != nil {
			return nil, fmt.Errorf("scan exercise image: %w", err)
		}
		images[variant] = image
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read exercise images: %w", err)
	}
	return images, nil
}

func (s *Store) exerciseSteps(id int64) ([]string, error) {
	rows, err := s.db.Query("SELECT instruction FROM exercise_steps WHERE exercise_id = ? ORDER BY position", id)
	if err != nil {
		return nil, fmt.Errorf("query exercise steps: %w", err)
	}
	defer rows.Close()
	var steps []string
	for rows.Next() {
		var step string
		if err := rows.Scan(&step); err != nil {
			return nil, fmt.Errorf("scan exercise step: %w", err)
		}
		steps = append(steps, step)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read exercise steps: %w", err)
	}
	return steps, nil
}
