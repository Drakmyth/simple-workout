package app

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestExerciseList(t *testing.T) {
	application := newTestApp(t)
	request := httptest.NewRequest(http.MethodGet, "/exercises", nil)
	response := httptest.NewRecorder()
	application.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	body := response.Body.String()
	if !strings.Contains(body, "Goblet squat") || !strings.Contains(body, "goblet-squat-male.webp") {
		t.Fatal("catalog did not contain the seeded exercise and selected image")
	}
	cookie := response.Result().Cookies()[0]
	if cookie.Name != alternatingCookie || cookie.Value != imageFemale {
		t.Fatalf("alternating cookie = (%q, %q), want (%q, %q)", cookie.Name, cookie.Value, alternatingCookie, imageFemale)
	}
}

func TestExerciseDetailUsesAlternatingCookie(t *testing.T) {
	application := newTestApp(t)
	request := httptest.NewRequest(http.MethodGet, "/exercises/dead-bug", nil)
	request.AddCookie(&http.Cookie{Name: alternatingCookie, Value: imageFemale})
	response := httptest.NewRecorder()
	application.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "dead-bug-female.webp") {
		t.Fatalf("detail response = (%d, %q), want female dead bug", response.Code, response.Body.String())
	}
	if got := response.Result().Cookies()[0].Value; got != imageMale {
		t.Fatalf("next alternating variant = %q, want %q", got, imageMale)
	}
}

func TestExerciseDetailNotFound(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/exercises/missing", nil)
	response := httptest.NewRecorder()
	newTestApp(t).Handler().ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestUpdateImagePreference(t *testing.T) {
	application := newTestApp(t)
	form := url.Values{"image_preference": {imageFemale}, "return_to": {"/exercises/dead-bug"}}
	request := httptest.NewRequest(http.MethodPost, "/settings/image-preference", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	application.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusSeeOther || response.Header().Get("Location") != "/exercises/dead-bug" {
		t.Fatalf("response = (%d, %q), want redirect to detail", response.Code, response.Header().Get("Location"))
	}
	preference, err := application.imagePreference()
	if err != nil || preference != imageFemale {
		t.Fatalf("preference = (%q, %v), want (%q, nil)", preference, err, imageFemale)
	}
}

func TestRejectInvalidImagePreference(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/settings/image-preference", strings.NewReader("image_preference=invalid"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	newTestApp(t).Handler().ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}
