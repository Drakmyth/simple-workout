package app

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

const guideVariantCookie = "exercise_guide_variant"

type exerciseView struct {
	Exercise
	Image ExerciseImage
}

type catalogPage struct {
	Exercises []exerciseView
	Variant   string
}

type exercisePage struct {
	Exercise exerciseView
	Variant  string
}

func (a *App) exerciseList(w http.ResponseWriter, r *http.Request) {
	exercises, err := a.exercises()
	if err != nil {
		a.serverError(w, "load exercise catalog", err)
		return
	}
	variant := guideVariant(r)
	views := make([]exerciseView, 0, len(exercises))
	for _, exercise := range exercises {
		views = append(views, exerciseView{Exercise: exercise, Image: exercise.Images[variant]})
	}
	a.render(w, "exercises.html", catalogPage{Exercises: views, Variant: variant})
}

func (a *App) exerciseDetail(w http.ResponseWriter, r *http.Request) {
	exercise, err := a.exercise(r.PathValue("slug"))
	if isNotFound(err) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, "load exercise", err)
		return
	}
	variant := guideVariant(r)
	a.render(w, "exercise.html", exercisePage{
		Exercise: exerciseView{Exercise: exercise, Image: exercise.Images[variant]},
		Variant:  variant,
	})
}

func (a *App) updateGuideVariant(w http.ResponseWriter, r *http.Request) {
	variant := r.FormValue("image_variant")
	if variant != imageMale && variant != imageFemale {
		http.Error(w, "Invalid image variant", http.StatusBadRequest)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     guideVariantCookie,
		Value:    variant,
		Path:     "/",
		MaxAge:   int((365 * 24 * time.Hour).Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	returnTo := r.FormValue("return_to")
	if !strings.HasPrefix(returnTo, "/") || strings.HasPrefix(returnTo, "//") {
		returnTo = "/exercises"
	}
	http.Redirect(w, r, returnTo, http.StatusSeeOther)
}

func guideVariant(r *http.Request) string {
	cookie, err := r.Cookie(guideVariantCookie)
	if err == nil && cookie.Value == imageFemale {
		return imageFemale
	}
	return imageMale
}

func (a *App) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := a.templates.ExecuteTemplate(w, name, data); err != nil {
		a.logger.Error("render page", "template", name, "error", err)
	}
}

func (a *App) serverError(w http.ResponseWriter, operation string, err error) {
	a.logger.Error(operation, "error", err)
	http.Error(w, fmt.Sprintf("%s failed", operation), http.StatusInternalServerError)
}
