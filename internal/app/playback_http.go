package app

import (
	"net/http"
	"strconv"
	"strings"
)

type playbackExerciseView struct {
	PlaybackExercise
	Weight  string
	Variant string
	Image   ExerciseImage
}

type playbackSetupPage struct {
	Workout   PlaybackWorkout
	Exercises []playbackExerciseView
	Settings  PlaybackSettings
}

type playbackSlide struct {
	Kind          string
	Unit          string
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
	Current         *playbackSlide
	Exercises       []playbackExerciseView
	Step            int
	NextStep        int
	Complete        bool
	RestSeconds     int
	ImagePreference string
	SaveWeights     bool
	SavedWeights    bool
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
	if err == nil {
		err = assignPlaybackVariants(views, preference, r.Form["variant"])
	}
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
	step := 0
	if value := r.FormValue("step"); value != "" {
		step, err = strconv.Atoi(value)
	}
	slides := buildPlaybackSlides(views, restSeconds, unit)
	if err != nil || step < 0 || step > len(slides) {
		http.Error(w, "Invalid playback step", http.StatusBadRequest)
		return
	}
	page := playbackPage{Workout: workout, Slides: slides, Exercises: views, Step: step, NextStep: step + 1,
		Complete: step == len(slides), RestSeconds: restSeconds, ImagePreference: preference,
		SaveWeights: r.FormValue("save_weights") == "true"}
	if page.Complete && page.SaveWeights {
		weights, err := playbackWeights(views)
		if err == nil {
			err = a.store.saveWorkoutWeights(id, weights)
		}
		if err != nil {
			a.serverError(w, "save workout weights", err)
			return
		}
		page.SavedWeights = true
	} else if !page.Complete {
		page.Current = &slides[step]
	}
	a.render(w, "playback.html", page)
}

func (a *App) playbackComplete(w http.ResponseWriter, r *http.Request) {
	id, ok := pathInt(w, r, "id")
	if !ok {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid completion", http.StatusBadRequest)
		return
	}
	weights, err := playbackWeightValues(r.Form["weight"])
	if err != nil {
		http.Error(w, "Invalid completion", http.StatusBadRequest)
		return
	}
	if err := a.store.saveWorkoutWeights(id, weights); err != nil {
		a.serverError(w, "save workout weights", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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

func playbackWeights(exercises []playbackExerciseView) ([]*float64, error) {
	values := make([]string, len(exercises))
	for index, exercise := range exercises {
		values[index] = exercise.Weight
	}
	return playbackWeightValues(values)
}

func playbackWeightValues(values []string) ([]*float64, error) {
	weights := make([]*float64, len(values))
	for index, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		weight, err := strconv.ParseFloat(value, 64)
		if err != nil || weight < 0 {
			return nil, strconv.ErrSyntax
		}
		weights[index] = &weight
	}
	return weights, nil
}

func assignPlaybackVariants(exercises []playbackExerciseView, preference string, submitted []string) error {
	if len(submitted) > 0 && len(submitted) != len(exercises) {
		return strconv.ErrSyntax
	}
	next := ""
	for index := range exercises {
		variant := ""
		if len(submitted) > 0 {
			variant = submitted[index]
			if variant != imageMale && variant != imageFemale {
				return strconv.ErrSyntax
			}
		} else {
			var err error
			variant, next, err = selectImageVariant(preference, next)
			if err != nil {
				return err
			}
		}
		exercises[index].Variant = variant
		exercises[index].Image = exercises[index].Images[variant]
	}
	return nil
}

func buildPlaybackSlides(exercises []playbackExerciseView, restSeconds int, unit string) []playbackSlide {
	totalSets := 0
	for _, exercise := range exercises {
		totalSets += exercise.Sets
	}
	slides := make([]playbackSlide, 0, totalSets*2-1)
	completed := 0
	for exerciseIndex, exercise := range exercises {
		for set := 1; set <= exercise.Sets; set++ {
			completed++
			slides = append(slides, playbackSlide{Kind: "set", Unit: unit, Exercise: exercise, ExerciseIndex: exerciseIndex,
				SetNumber: set, CompletedSets: completed, TotalSets: totalSets})
			if completed < totalSets {
				slides = append(slides, playbackSlide{Kind: "rest", Unit: unit, RestSeconds: restSeconds,
					CompletedSets: completed, TotalSets: totalSets})
			}
		}
	}
	return slides
}
