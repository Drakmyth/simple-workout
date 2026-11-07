package app

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
