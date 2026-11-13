package app

import (
	"net/http"
	"strconv"
	"strings"
)

type playbackExerciseView struct {
	PlaybackExercise
	Index   int
	Note    string
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
	SetNumber     int
	CompletedSets int
	TotalSets     int
	RestSeconds   int
}

type weightSaveStatus string

const (
	playbackAlternatingCookie = "playback_image_next"

	weightSaveNone    weightSaveStatus = "none"
	weightSavePending weightSaveStatus = "pending"
	weightSaveDone    weightSaveStatus = "done"
)

type playbackPage struct {
	RestSound        bool
	Summary          bool
	Unit             string
	NoteExercise     *playbackExerciseView
	NoteIndex        int
	KeepScreenOn     bool
	Workout          PlaybackWorkout
	Current          *playbackSlide
	Exercises        []playbackExerciseView
	Step             int
	NextStep         int
	Complete         bool
	RestSeconds      int
	ImagePreference  string
	WeightSaveStatus weightSaveStatus
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
	keepScreenOn := r.FormValue("keep_screen_on") == "true"
	restSound := r.FormValue("rest_sound") == "true"
	views, err := parsePlaybackWeights(workout.Exercises, r.Form["weight"])
	variants := r.Form["variant"]
	initialStep := len(variants) == 0
	alternatingNext := ""
	if initialStep && preference == imageAlternating {
		if cookie, cookieErr := r.Cookie(playbackAlternatingCookie); cookieErr == nil {
			alternatingNext = cookie.Value
		}
	}
	if err == nil {
		alternatingNext, err = assignPlaybackVariants(views, preference, variants, alternatingNext)
	}
	if err != nil || restSeconds < 0 || !validImagePreference(preference) {
		http.Error(w, "Invalid playback setup", http.StatusBadRequest)
		return
	}
	if initialStep && preference == imageAlternating {
		http.SetCookie(w, &http.Cookie{Name: playbackAlternatingCookie, Value: alternatingNext, Path: "/",
			MaxAge: 31536000, HttpOnly: true, SameSite: http.SameSiteLaxMode})
	}
	if err := a.store.updatePlaybackSettings(restSeconds, preference, keepScreenOn, restSound); err != nil {
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
	if err != nil || step < 0 || step > len(slides)+1 {
		http.Error(w, "Invalid playback step", http.StatusBadRequest)
		return
	}
	for index := range views {
		note := r.FormValue("note_" + strconv.Itoa(index))
		if len(note) > 4000 {
			http.Error(w, "Performance note is too long", http.StatusBadRequest)
			return
		}
		views[index].Note = note
	}
	saveStatus := weightSaveNone
	if r.FormValue("save_weights") == "true" {
		saveStatus = weightSavePending
	}
	page := playbackPage{RestSound: restSound, KeepScreenOn: keepScreenOn, Workout: workout, Exercises: views, Step: step, NextStep: step + 1,
		Complete: step >= len(slides), RestSeconds: restSeconds, ImagePreference: preference,
		WeightSaveStatus: saveStatus, Summary: step == len(slides)+1, Unit: unit, NoteIndex: -1}
	if page.Summary && page.WeightSaveStatus == weightSavePending {
		page.WeightSaveStatus = weightSaveDone
	}
	if page.Complete && page.WeightSaveStatus == weightSavePending {
		weights, err := playbackWeights(views)
		if err == nil {
			err = a.store.saveWorkoutWeights(id, weights)
		}
		if err != nil {
			a.serverError(w, "save workout weights", err)
			return
		}
		page.WeightSaveStatus = weightSaveDone
	} else if !page.Complete {
		page.Current = &slides[step]
	}
	if page.Complete && !page.Summary {
		page.NoteIndex = len(views) - 1
	} else if !page.Complete && slides[step].Kind == "rest" {
		previous := slides[step-1]
		if previous.SetNumber == previous.Exercise.Sets {
			page.NoteIndex = previous.Exercise.Index
		}
	}
	if page.NoteIndex >= 0 {
		page.NoteExercise = &views[page.NoteIndex]
	}
	a.render(w, "playback.html", page)
}

func playbackExerciseViews(exercises []PlaybackExercise) []playbackExerciseView {
	views := make([]playbackExerciseView, len(exercises))
	for index, exercise := range exercises {
		weight := ""
		if exercise.Weight != nil {
			weight = strconv.FormatFloat(*exercise.Weight, 'f', -1, 64)
		}
		views[index] = playbackExerciseView{PlaybackExercise: exercise, Weight: weight, Index: index}
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

func assignPlaybackVariants(exercises []playbackExerciseView, preference string, submitted []string, next string) (string, error) {
	if len(submitted) > 0 && len(submitted) != len(exercises) {
		return next, strconv.ErrSyntax
	}
	for index := range exercises {
		variant := ""
		if len(submitted) > 0 {
			variant = submitted[index]
			if variant != imageMale && variant != imageFemale {
				return next, strconv.ErrSyntax
			}
		} else {
			var err error
			variant, next, err = selectImageVariant(preference, next)
			if err != nil {
				return next, err
			}
		}
		exercises[index].Variant = variant
		exercises[index].Image = exercises[index].Images[variant]
	}
	return next, nil
}

func buildPlaybackSlides(exercises []playbackExerciseView, restSeconds int, unit string) []playbackSlide {
	totalSets := 0
	for _, exercise := range exercises {
		totalSets += exercise.Sets
	}
	slides := make([]playbackSlide, 0, totalSets*2-1)
	completed := 0
	for _, exercise := range exercises {
		for set := 1; set <= exercise.Sets; set++ {
			completed++
			slides = append(slides, playbackSlide{Kind: "set", Unit: unit, Exercise: exercise,
				SetNumber: set, CompletedSets: completed, TotalSets: totalSets})
			if completed < totalSets {
				slides = append(slides, playbackSlide{Kind: "rest", Unit: unit, RestSeconds: restSeconds,
					CompletedSets: completed, TotalSets: totalSets})
			}
		}
	}
	for index := range slides {
		if slides[index].Kind == "rest" {
			next := slides[index+1]
			slides[index].Exercise = next.Exercise
			slides[index].SetNumber = next.SetNumber
		}
	}
	return slides
}
