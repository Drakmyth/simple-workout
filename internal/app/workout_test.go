package app

import "testing"

func TestSeededWorkouts(t *testing.T) {
	application := newTestApp(t)
	workouts, err := application.workouts()
	if err != nil {
		t.Fatalf("workouts() error = %v", err)
	}
	if len(workouts) != 2 || workouts[0].Name != "Workout A" || workouts[0].ExerciseCount != 5 {
		t.Fatalf("workouts = %#v, want Workout A and Workout B with five exercises", workouts)
	}
	workout, err := application.workout(2)
	if err != nil {
		t.Fatalf("workout(2) error = %v", err)
	}
	if len(workout.Exercises) != 5 || workout.Exercises[2].Slug != "one-arm-dumbbell-row" {
		t.Fatalf("Workout B exercises = %#v", workout.Exercises)
	}
	if workout.Exercises[1].Weight != nil || workout.Exercises[2].Weight == nil || *workout.Exercises[2].Weight != 3 {
		t.Fatal("Workout B bodyweight and row weights were not seeded correctly")
	}
	if _, err := application.workout(99); !isNotFound(err) {
		t.Fatalf("workout(99) error = %v, want not found", err)
	}
}

func TestRenameAndUpdateWorkout(t *testing.T) {
	application := newTestApp(t)
	if err := application.renameWorkout(1, "  Strength day  "); err != nil {
		t.Fatalf("renameWorkout() error = %v", err)
	}
	weight := 4.0
	if err := application.updateWorkoutExercise(1, 0, 3, 8, &weight); err != nil {
		t.Fatalf("updateWorkoutExercise() error = %v", err)
	}
	workout, err := application.workout(1)
	if err != nil {
		t.Fatalf("workout(1) error = %v", err)
	}
	first := workout.Exercises[0]
	if workout.Name != "Strength day" || first.Sets != 3 || first.Reps != 8 || first.Weight == nil || *first.Weight != 4 {
		t.Fatalf("updated workout = %#v", workout)
	}
	if err := application.renameWorkout(1, " "); err == nil {
		t.Fatal("renameWorkout(empty) error = nil")
	}
	if err := application.updateWorkoutExercise(1, 0, 0, 8, nil); err == nil {
		t.Fatal("updateWorkoutExercise(invalid) error = nil")
	}
}

func TestAddMoveAndRemoveWorkoutExercise(t *testing.T) {
	application := newTestApp(t)
	if err := application.addWorkoutExercise(1, 9); err != nil {
		t.Fatalf("addWorkoutExercise() error = %v", err)
	}
	if err := application.moveWorkoutExercise(1, 5, -1); err != nil {
		t.Fatalf("moveWorkoutExercise() error = %v", err)
	}
	workout, err := application.workout(1)
	if err != nil {
		t.Fatalf("workout(1) error = %v", err)
	}
	if workout.Exercises[4].Slug != "standing-calf-raise" || workout.Exercises[5].Slug != "dead-bug" {
		t.Fatalf("moved order = %#v", workout.Exercises)
	}
	if err := application.removeWorkoutExercise(1, 4); err != nil {
		t.Fatalf("removeWorkoutExercise() error = %v", err)
	}
	workout, err = application.workout(1)
	if err != nil {
		t.Fatalf("workout(1) after remove error = %v", err)
	}
	if len(workout.Exercises) != 5 {
		t.Fatalf("exercise count = %d, want 5", len(workout.Exercises))
	}
	for position, exercise := range workout.Exercises {
		if exercise.Position != position {
			t.Fatalf("exercise position = %d, want %d", exercise.Position, position)
		}
	}
	if err := application.addWorkoutExercise(1, 5); err != nil {
		t.Fatalf("add duplicate exercise error = %v", err)
	}
	workout, err = application.workout(1)
	if err != nil || len(workout.Exercises) != 6 || workout.Exercises[5].Slug != "dead-bug" {
		t.Fatalf("workout with duplicate exercise = (%#v, %v)", workout, err)
	}
}
