package app

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
