package app

import (
	"database/sql"
	"fmt"
)

func (s *Store) playbackWorkout(id int64) (PlaybackWorkout, error) {
	var workout PlaybackWorkout
	if err := s.db.QueryRow("SELECT id, name FROM workouts WHERE id = ?", id).Scan(&workout.ID, &workout.Name); err != nil {
		return PlaybackWorkout{}, fmt.Errorf("query playback workout: %w", err)
	}
	rows, err := s.db.Query(`SELECT e.id, e.slug, e.name, e.notes, we.sets, we.reps, we.weight,
		MAX(CASE WHEN i.variant = 'male' THEN i.path END),
		MAX(CASE WHEN i.variant = 'male' THEN i.alt_text END),
		MAX(CASE WHEN i.variant = 'female' THEN i.path END),
		MAX(CASE WHEN i.variant = 'female' THEN i.alt_text END)
		FROM workout_exercises we
		JOIN exercises e ON e.id = we.exercise_id
		JOIN exercise_images i ON i.exercise_id = e.id
		WHERE we.workout_id = ?
		GROUP BY we.position, e.id
		ORDER BY we.position`, id)
	if err != nil {
		return PlaybackWorkout{}, fmt.Errorf("query playback exercises: %w", err)
	}
	var exerciseIDs []int64
	for rows.Next() {
		var exerciseID int64
		var exercise PlaybackExercise
		var weight sql.NullFloat64
		var male, female ExerciseImage
		if err := rows.Scan(&exerciseID, &exercise.Slug, &exercise.Name, &exercise.Notes,
			&exercise.Sets, &exercise.Reps, &weight,
			&male.Path, &male.AltText, &female.Path, &female.AltText); err != nil {
			return PlaybackWorkout{}, fmt.Errorf("scan playback exercise: %w", err)
		}
		if weight.Valid {
			exercise.Weight = &weight.Float64
		}
		exercise.Images = map[string]ExerciseImage{imageMale: male, imageFemale: female}
		exerciseIDs = append(exerciseIDs, exerciseID)
		workout.Exercises = append(workout.Exercises, exercise)
	}
	if err := rows.Close(); err != nil {
		return PlaybackWorkout{}, fmt.Errorf("close playback exercises: %w", err)
	}
	if err := rows.Err(); err != nil {
		return PlaybackWorkout{}, fmt.Errorf("read playback exercises: %w", err)
	}
	stepCache := make(map[int64][]string)
	for index, exerciseID := range exerciseIDs {
		steps, ok := stepCache[exerciseID]
		if !ok {
			steps, err = s.exerciseSteps(exerciseID)
			if err != nil {
				return PlaybackWorkout{}, err
			}
			stepCache[exerciseID] = steps
		}
		workout.Exercises[index].Steps = steps
	}
	return workout, nil
}

func (s *Store) saveWorkoutWeights(workoutID int64, weights []*float64) error {
	for _, weight := range weights {
		if weight != nil && *weight < 0 {
			return fmt.Errorf("invalid workout weight")
		}
	}
	return s.withWorkoutTx(workoutID, func(tx *sql.Tx) error {
		var count int
		if err := tx.QueryRow("SELECT COUNT(*) FROM workout_exercises WHERE workout_id = ?", workoutID).Scan(&count); err != nil {
			return fmt.Errorf("count workout exercises: %w", err)
		}
		if count != len(weights) {
			return fmt.Errorf("weight count %d does not match exercise count %d", len(weights), count)
		}
		for position, weight := range weights {
			result, err := tx.Exec("UPDATE workout_exercises SET weight = ? WHERE workout_id = ? AND position = ?", weight, workoutID, position)
			if err != nil {
				return fmt.Errorf("save workout weight: %w", err)
			}
			if err := requireChanged(result, "workout exercise"); err != nil {
				return err
			}
		}
		return nil
	})
}
