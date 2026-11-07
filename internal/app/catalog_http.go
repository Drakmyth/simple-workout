package app

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

const alternatingCookie = "exercise_image_next"

type exerciseView struct {
	Exercise
	Image ExerciseImage
}

type catalogPage struct {
	Exercises  []exerciseView
	Preference string
}

type exercisePage struct {
	Exercise   exerciseView
	Preference string
}

func (a *App) exerciseList(w http.ResponseWriter, r *http.Request) {
	exercises, err := a.exercises()
	if err != nil {
		a.serverError(w, "load exercise catalog", err)
		return
	}
	preference, err := a.imagePreference()
	if err != nil {
		a.serverError(w, "load image preference", err)
		return
	}

	next := alternatingNext(r)
	views := make([]exerciseView, 0, len(exercises))
	for _, exercise := range exercises {
		variant, updatedNext, err := selectImageVariant(preference, next)
		if err != nil {
			a.serverError(w, "select exercise image", err)
			return
		}
		next = updatedNext
		views = append(views, exerciseView{Exercise: exercise, Image: exercise.Images[variant]})
	}
	if preference == imageAlternating {
		setAlternatingNext(w, next)
	}
	a.render(w, "exercises.html", catalogPage{Exercises: views, Preference: preference})
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
	preference, err := a.imagePreference()
	if err != nil {
		a.serverError(w, "load image preference", err)
		return
	}
	variant, next, err := selectImageVariant(preference, alternatingNext(r))
	if err != nil {
		a.serverError(w, "select exercise image", err)
		return
	}
	if preference == imageAlternating {
		setAlternatingNext(w, next)
	}
	a.render(w, "exercise.html", exercisePage{
		Exercise:   exerciseView{Exercise: exercise, Image: exercise.Images[variant]},
		Preference: preference,
	})
}

func (a *App) updateImagePreference(w http.ResponseWriter, r *http.Request) {
	preference := r.FormValue("image_preference")
	if err := a.setImagePreference(preference); err != nil {
		http.Error(w, "Invalid image preference", http.StatusBadRequest)
		return
	}
	if preference == imageAlternating {
		http.SetCookie(w, &http.Cookie{Name: alternatingCookie, Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode})
	}
	returnTo := r.FormValue("return_to")
	if !strings.HasPrefix(returnTo, "/") || strings.HasPrefix(returnTo, "//") {
		returnTo = "/exercises"
	}
	http.Redirect(w, r, returnTo, http.StatusSeeOther)
}

func alternatingNext(r *http.Request) string {
	cookie, err := r.Cookie(alternatingCookie)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func setAlternatingNext(w http.ResponseWriter, next string) {
	http.SetCookie(w, &http.Cookie{
		Name:     alternatingCookie,
		Value:    next,
		Path:     "/",
		MaxAge:   int((365 * 24 * time.Hour).Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
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
