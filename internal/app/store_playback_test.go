package app

import "testing"

func TestPlaybackWorkout(t *testing.T) {
	application := newTestApp(t)
	workout, err := application.store.playbackWorkout(1)
	if err != nil {
		t.Fatalf("playbackWorkout() error = %v", err)
	}
	if workout.Name != "Workout A" || len(workout.Exercises) != 5 {
		t.Fatalf("playback workout = %#v", workout)
	}
	first := workout.Exercises[0]
	if first.Slug != "goblet-squat" || first.Sets != 2 || first.Reps != 10 || len(first.Steps) != 3 {
		t.Fatalf("first playback exercise = %#v", first)
	}
	if first.Images[imageMale].Path == "" || first.Images[imageFemale].Path == "" {
		t.Fatal("playback exercise is missing an image variant")
	}
	if _, err := application.store.playbackWorkout(99); !isNotFound(err) {
		t.Fatalf("playbackWorkout(99) error = %v, want not found", err)
	}
}

func TestPlaybackSettings(t *testing.T) {
	application := newTestApp(t)
	settings, err := application.store.playbackSettings()
	if err != nil {
		t.Fatalf("playbackSettings() error = %v", err)
	}
	if settings.RestSeconds != 60 || settings.WeightUnit != "kg" || settings.ImagePreference != imageAlternating {
		t.Fatalf("initial playback settings = %#v", settings)
	}
	if err := application.store.updatePlaybackSettings(45, imageRandom); err != nil {
		t.Fatalf("updatePlaybackSettings() error = %v", err)
	}
	settings, err = application.store.playbackSettings()
	if err != nil || settings.RestSeconds != 45 || settings.ImagePreference != imageRandom {
		t.Fatalf("updated playback settings = (%#v, %v)", settings, err)
	}
	if err := application.store.updatePlaybackSettings(-1, imageMale); err == nil {
		t.Fatal("updatePlaybackSettings(invalid) error = nil")
	}
}

func TestSaveWorkoutWeights(t *testing.T) {
	application := newTestApp(t)
	weights := []*float64{floatPointer(4), floatPointer(5), floatPointer(6), floatPointer(7), nil}
	if err := application.store.saveWorkoutWeights(1, weights); err != nil {
		t.Fatalf("saveWorkoutWeights() error = %v", err)
	}
	workout, err := application.store.workout(1)
	if err != nil {
		t.Fatalf("workout(1) error = %v", err)
	}
	for index, want := range []float64{4, 5, 6, 7} {
		if workout.Exercises[index].Weight == nil || *workout.Exercises[index].Weight != want {
			t.Errorf("weight %d = %v, want %v", index, workout.Exercises[index].Weight, want)
		}
	}
	if workout.Exercises[4].Weight != nil {
		t.Fatalf("bodyweight exercise weight = %v, want nil", *workout.Exercises[4].Weight)
	}
}

func TestSaveWorkoutWeightsRejectsIncompleteBatch(t *testing.T) {
	application := newTestApp(t)
	if err := application.store.saveWorkoutWeights(1, []*float64{floatPointer(10)}); err == nil {
		t.Fatal("saveWorkoutWeights(incomplete) error = nil")
	}
	workout, err := application.store.workout(1)
	if err != nil || workout.Exercises[0].Weight == nil || *workout.Exercises[0].Weight != 2.5 {
		t.Fatalf("workout changed after rejected batch = (%#v, %v)", workout, err)
	}
}

func floatPointer(value float64) *float64 {
	return &value
}
