package app

import (
	"database/sql"
	"fmt"
	"strings"
)

type WorkoutSummary struct {
	ID            int64
	Name          string
	ExerciseCount int
}

type Workout struct {
	ID        int64
	Name      string
	Exercises []WorkoutExercise
}

type WorkoutExercise struct {
	Slug   string
	Name   string
	Sets   int
	Reps   int
	Weight *float64
}

func (a *App) weightUnit() (string, error) {
	var unit string
	if err := a.db.QueryRow("SELECT weight_unit FROM settings WHERE id = 1").Scan(&unit); err != nil {
		return "", fmt.Errorf("query weight unit: %w", err)
	}
	return unit, nil
}

func (a *App) workouts() ([]WorkoutSummary, error) {
	rows, err := a.db.Query(`SELECT w.id, w.name, COUNT(we.exercise_id)
		FROM workouts w LEFT JOIN workout_exercises we ON we.workout_id = w.id
		GROUP BY w.id ORDER BY w.id`)
	if err != nil {
		return nil, fmt.Errorf("query workouts: %w", err)
	}
	defer rows.Close()
	var workouts []WorkoutSummary
	for rows.Next() {
		var workout WorkoutSummary
		if err := rows.Scan(&workout.ID, &workout.Name, &workout.ExerciseCount); err != nil {
			return nil, fmt.Errorf("scan workout: %w", err)
		}
		workouts = append(workouts, workout)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read workouts: %w", err)
	}
	return workouts, nil
}

func (a *App) workout(id int64) (Workout, error) {
	var workout Workout
	if err := a.db.QueryRow("SELECT id, name FROM workouts WHERE id = ?", id).Scan(&workout.ID, &workout.Name); err != nil {
		return Workout{}, fmt.Errorf("query workout: %w", err)
	}
	rows, err := a.db.Query(`SELECT e.slug, e.name, we.sets, we.reps, we.weight
		FROM workout_exercises we JOIN exercises e ON e.id = we.exercise_id
		WHERE we.workout_id = ? ORDER BY we.position`, id)
	if err != nil {
		return Workout{}, fmt.Errorf("query workout exercises: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item WorkoutExercise
		var weight sql.NullFloat64
		if err := rows.Scan(&item.Slug, &item.Name, &item.Sets, &item.Reps, &weight); err != nil {
			return Workout{}, fmt.Errorf("scan workout exercise: %w", err)
		}
		if weight.Valid {
			item.Weight = &weight.Float64
		}
		workout.Exercises = append(workout.Exercises, item)
	}
	if err := rows.Err(); err != nil {
		return Workout{}, fmt.Errorf("read workout exercises: %w", err)
	}
	return workout, nil
}

func (a *App) renameWorkout(id int64, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("workout name is required")
	}
	result, err := a.db.Exec("UPDATE workouts SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", name, id)
	if err != nil {
		return fmt.Errorf("rename workout: %w", err)
	}
	return requireChanged(result, "workout")
}

func (a *App) updateWorkoutExercise(workoutID int64, position, sets, reps int, weight *float64) error {
	if position < 0 || sets < 1 || reps < 1 || weight != nil && *weight < 0 {
		return fmt.Errorf("invalid workout exercise values")
	}
	return a.withWorkoutTx(workoutID, func(tx *sql.Tx) error {
		result, err := tx.Exec(`UPDATE workout_exercises SET sets = ?, reps = ?, weight = ?
			WHERE workout_id = ? AND position = ?`, sets, reps, weight, workoutID, position)
		if err != nil {
			return fmt.Errorf("update workout exercise: %w", err)
		}
		return requireChanged(result, "workout exercise")
	})
}

func (a *App) addWorkoutExercise(workoutID, exerciseID int64) error {
	return a.withWorkoutTx(workoutID, func(tx *sql.Tx) error {
		result, err := tx.Exec(`INSERT INTO workout_exercises (workout_id, position, exercise_id, sets, reps)
			SELECT ?, COALESCE(MAX(position), -1) + 1, ?, 2, 10 FROM workout_exercises
			WHERE workout_id = ?`, workoutID, exerciseID, workoutID)
		if err != nil {
			return fmt.Errorf("add workout exercise: %w", err)
		}
		return requireChanged(result, "available exercise")
	})
}

func (a *App) removeWorkoutExercise(workoutID int64, position int) error {
	if position < 0 {
		return fmt.Errorf("invalid workout exercise position")
	}
	return a.withWorkoutTx(workoutID, func(tx *sql.Tx) error {
		result, err := tx.Exec("DELETE FROM workout_exercises WHERE workout_id = ? AND position = ?", workoutID, position)
		if err != nil {
			return fmt.Errorf("remove workout exercise: %w", err)
		}
		if err := requireChanged(result, "workout exercise"); err != nil {
			return err
		}
		var offset int
		if err := tx.QueryRow("SELECT COALESCE(MAX(position), 0) + 1 FROM workout_exercises WHERE workout_id = ?", workoutID).Scan(&offset); err != nil {
			return fmt.Errorf("find workout position offset: %w", err)
		}
		if _, err := tx.Exec("UPDATE workout_exercises SET position = position + ? WHERE workout_id = ? AND position > ?", offset, workoutID, position); err != nil {
			return fmt.Errorf("stage workout positions: %w", err)
		}
		if _, err := tx.Exec("UPDATE workout_exercises SET position = position - ? WHERE workout_id = ? AND position >= ?", offset+1, workoutID, position+1+offset); err != nil {
			return fmt.Errorf("compact workout positions: %w", err)
		}
		return nil
	})
}

func (a *App) moveWorkoutExercise(workoutID int64, position, direction int) error {
	target := position + direction
	if position < 0 || target < 0 || direction != -1 && direction != 1 {
		return fmt.Errorf("invalid workout exercise move")
	}
	return a.withWorkoutTx(workoutID, func(tx *sql.Tx) error {
		var temporary int
		if err := tx.QueryRow("SELECT COALESCE(MAX(position), -1) + 1 FROM workout_exercises WHERE workout_id = ?", workoutID).Scan(&temporary); err != nil {
			return fmt.Errorf("find temporary workout position: %w", err)
		}
		result, err := tx.Exec("UPDATE workout_exercises SET position = ? WHERE workout_id = ? AND position = ?", temporary, workoutID, position)
		if err != nil {
			return fmt.Errorf("stage workout exercise move: %w", err)
		}
		if err := requireChanged(result, "workout exercise"); err != nil {
			return err
		}
		result, err = tx.Exec("UPDATE workout_exercises SET position = ? WHERE workout_id = ? AND position = ?", position, workoutID, target)
		if err != nil {
			return fmt.Errorf("move adjacent workout exercise: %w", err)
		}
		if err := requireChanged(result, "adjacent workout exercise"); err != nil {
			return err
		}
		_, err = tx.Exec("UPDATE workout_exercises SET position = ? WHERE workout_id = ? AND position = ?", target, workoutID, temporary)
		return err
	})
}

func (a *App) withWorkoutTx(workoutID int64, operation func(*sql.Tx) error) error {
	tx, err := a.db.Begin()
	if err != nil {
		return fmt.Errorf("begin workout update: %w", err)
	}
	if err := operation(tx); err != nil {
		tx.Rollback()
		return err
	}
	if _, err := tx.Exec("UPDATE workouts SET updated_at = CURRENT_TIMESTAMP WHERE id = ?", workoutID); err != nil {
		tx.Rollback()
		return fmt.Errorf("touch workout: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit workout update: %w", err)
	}
	return nil
}

func requireChanged(result sql.Result, subject string) error {
	changed, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("count changed %s: %w", subject, err)
	}
	if changed == 0 {
		return fmt.Errorf("%s not found", subject)
	}
	return nil
}
