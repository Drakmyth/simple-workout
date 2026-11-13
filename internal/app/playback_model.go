package app

type PlaybackSettings struct {
	KeepImageVisible bool
	RestSound        bool
	KeepScreenOn     bool
	RestSeconds      int
	WeightUnit       string
	ImagePreference  string
}

type PlaybackWorkout struct {
	ID        int64
	Name      string
	Exercises []PlaybackExercise
}

type PlaybackExercise struct {
	Slug   string
	Name   string
	Notes  string
	Steps  []string
	Images map[string]ExerciseImage
	Sets   int
	Reps   int
	Weight *float64
}
