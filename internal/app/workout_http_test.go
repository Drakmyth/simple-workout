package app

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestWorkoutPages(t *testing.T) {
	application := newTestApp(t)
	for path, content := range map[string]string{
		"/workouts":        "Workout A",
		"/workouts/1/edit": "Goblet squat",
	} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		application.Handler().ServeHTTP(response, request)
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), content) {
			t.Fatalf("GET %s = (%d, %q), want content %q", path, response.Code, response.Body.String(), content)
		}
	}
}

func TestWorkoutEditNotFound(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/workouts/99/edit", nil)
	response := httptest.NewRecorder()
	newTestApp(t).Handler().ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestWorkoutRenameAndTargetUpdate(t *testing.T) {
	application := newTestApp(t)
	postWorkoutForm(t, application, "/workouts/1/rename", url.Values{"name": {"Day one"}})
	postWorkoutForm(t, application, "/workouts/1/exercises/0/update", url.Values{
		"sets": {"3"}, "reps": {"8"}, "weight": {"4.5"},
	})
	workout, err := application.store.workout(1)
	if err != nil {
		t.Fatalf("workout(1) error = %v", err)
	}
	first := workout.Exercises[0]
	if workout.Name != "Day one" || first.Sets != 3 || first.Reps != 8 || first.Weight == nil || *first.Weight != 4.5 {
		t.Fatalf("updated workout = %#v", workout)
	}
}

func TestWorkoutAddMoveAndRemoveRoutes(t *testing.T) {
	application := newTestApp(t)
	postWorkoutForm(t, application, "/workouts/1/exercises", url.Values{"exercise_id": {"9"}})
	postWorkoutForm(t, application, "/workouts/1/exercises/5/move", url.Values{"direction": {"up"}})
	postWorkoutForm(t, application, "/workouts/1/exercises/4/remove", nil)
	workout, err := application.store.workout(1)
	if err != nil || len(workout.Exercises) != 5 || workout.Exercises[4].Slug != "dead-bug" {
		t.Fatalf("edited workout = (%#v, %v)", workout, err)
	}
}

func TestWorkoutRouteRejectsInvalidInput(t *testing.T) {
	application := newTestApp(t)
	request := httptest.NewRequest(http.MethodPost, "/workouts/1/exercises/0/update", strings.NewReader("sets=0&reps=10"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	application.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func postWorkoutForm(t *testing.T, application *App, path string, values url.Values) {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(values.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	application.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusSeeOther || response.Header().Get("Location") != "/workouts/1/edit" {
		t.Fatalf("POST %s = (%d, %q), want edit redirect", path, response.Code, response.Header().Get("Location"))
	}
}
