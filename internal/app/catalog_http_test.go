package app

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestExerciseListDefaultsToMale(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/exercises", nil)
	response := httptest.NewRecorder()
	newTestApp(t).Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	body := response.Body.String()
	if !strings.Contains(body, "Goblet squat") || !strings.Contains(body, "goblet-squat-male.webp") {
		t.Fatal("catalog did not contain the seeded exercise and default male image")
	}
}

func TestExerciseDetailUsesGuideVariantCookie(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/exercises/dead-bug", nil)
	request.AddCookie(&http.Cookie{Name: guideVariantCookie, Value: imageFemale})
	response := httptest.NewRecorder()
	newTestApp(t).Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "dead-bug-female.webp") {
		t.Fatalf("detail response = (%d, %q), want female dead bug", response.Code, response.Body.String())
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

func TestUpdateGuideVariant(t *testing.T) {
	form := url.Values{"image_variant": {imageFemale}, "return_to": {"/exercises/dead-bug"}}
	request := httptest.NewRequest(http.MethodPost, "/exercises/image-variant", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	newTestApp(t).Handler().ServeHTTP(response, request)

	if response.Code != http.StatusSeeOther || response.Header().Get("Location") != "/exercises/dead-bug" {
		t.Fatalf("response = (%d, %q), want redirect to detail", response.Code, response.Header().Get("Location"))
	}
	cookie := response.Result().Cookies()[0]
	if cookie.Name != guideVariantCookie || cookie.Value != imageFemale {
		t.Fatalf("guide cookie = (%q, %q), want (%q, %q)", cookie.Name, cookie.Value, guideVariantCookie, imageFemale)
	}
}

func TestRejectInvalidGuideVariant(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/exercises/image-variant", strings.NewReader("image_variant=random"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	newTestApp(t).Handler().ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}
