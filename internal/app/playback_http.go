package app

import (
	"net/http"
	"strconv"
	"strings"
)

type playbackExerciseView struct {
	PlaybackExercise
	Weight string
}

type playbackSetupPage struct {
	Workout   PlaybackWorkout
	Exercises []playbackExerciseView
	Settings  PlaybackSettings
}

type playbackSlide struct {
	Kind          string
	Exercise      playbackExerciseView
	ExerciseIndex int
	SetNumber     int
	CompletedSets int
	TotalSets     int
	RestSeconds   int
}

type playbackPage struct {
	Workout         PlaybackWorkout
	Slides          []playbackSlide
	ImagePreference string
	Unit            string
}

func (a *App) playbackSetup(w http.ResponseWriter, r *http.Request) {
	id, ok := pathInt(w, r, "id")
	if !ok {
		return
	}
	workout, err := a.store.playbackWorkout(id)
	if isNotFound(err) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, "load workout playback", err)
		return
	}
	settings, err := a.store.playbackSettings()
	if err != nil {
		a.serverError(w, "load playback settings", err)
		return
	}
	a.render(w, "playback-setup.html", playbackSetupPage{
		Workout: workout, Exercises: playbackExerciseViews(workout.Exercises), Settings: settings,
	})
}

func (a *App) playbackStart(w http.ResponseWriter, r *http.Request) {
	id, ok := pathInt(w, r, "id")
	if !ok {
		return
	}
	workout, err := a.store.playbackWorkout(id)
	if isNotFound(err) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, "load workout playback", err)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid playback setup", http.StatusBadRequest)
		return
	}
	restSeconds, err := strconv.Atoi(r.FormValue("rest_seconds"))
	preference := r.FormValue("image_preference")
	views, err := parsePlaybackWeights(workout.Exercises, r.Form["weight"])
	if err != nil || restSeconds < 0 || !validImagePreference(preference) {
		http.Error(w, "Invalid playback setup", http.StatusBadRequest)
		return
	}
	if err := a.store.updatePlaybackSettings(restSeconds, preference); err != nil {
		a.serverError(w, "update playback settings", err)
		return
	}
	unit, err := a.store.weightUnit()
	if err != nil {
		a.serverError(w, "load weight unit", err)
		return
	}
	a.render(w, "playback.html", playbackPage{
		Workout: workout, Slides: buildPlaybackSlides(views, restSeconds), ImagePreference: preference, Unit: unit,
	})
}

func playbackExerciseViews(exercises []PlaybackExercise) []playbackExerciseView {
	views := make([]playbackExerciseView, len(exercises))
	for index, exercise := range exercises {
		weight := ""
		if exercise.Weight != nil {
			weight = strconv.FormatFloat(*exercise.Weight, 'f', -1, 64)
		}
		views[index] = playbackExerciseView{PlaybackExercise: exercise, Weight: weight}
	}
	return views
}

func parsePlaybackWeights(exercises []PlaybackExercise, values []string) ([]playbackExerciseView, error) {
	if len(values) != len(exercises) {
		return nil, strconv.ErrSyntax
	}
	views := playbackExerciseViews(exercises)
	for index, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		weight, err := strconv.ParseFloat(value, 64)
		if err != nil || weight < 0 {
			return nil, strconv.ErrSyntax
		}
		views[index].Weight = strconv.FormatFloat(weight, 'f', -1, 64)
	}
	return views, nil
}

func buildPlaybackSlides(exercises []playbackExerciseView, restSeconds int) []playbackSlide {
	totalSets := 0
	for _, exercise := range exercises {
		totalSets += exercise.Sets
	}
	slides := make([]playbackSlide, 0, totalSets*2-1)
	completed := 0
	for exerciseIndex, exercise := range exercises {
		for set := 1; set <= exercise.Sets; set++ {
			completed++
			slides = append(slides, playbackSlide{Kind: "set", Exercise: exercise, ExerciseIndex: exerciseIndex,
				SetNumber: set, CompletedSets: completed, TotalSets: totalSets})
			if completed < totalSets {
				slides = append(slides, playbackSlide{Kind: "rest", RestSeconds: restSeconds,
					CompletedSets: completed, TotalSets: totalSets})
			}
		}
	}
	return slides
}
