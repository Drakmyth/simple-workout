package app

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestPlaybackSetup(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/workouts/1/start", nil)
	response := httptest.NewRecorder()
	newTestApp(t).Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	body := response.Body.String()
	if !strings.Contains(body, "Start workout") || !strings.Contains(body, "Goblet squat") || !strings.Contains(body, "value=\"2.5\"") {
		t.Fatal("playback setup did not contain workout defaults")
	}
}

func TestPlaybackStart(t *testing.T) {
	application := newTestApp(t)
	form := url.Values{
		"rest_seconds":     {"30"},
		"image_preference": {imageRandom},
		"weight":           {"3", "4", "5", "6", ""},
	}
	request := httptest.NewRequest(http.MethodPost, "/workouts/1/play", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	application.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	body := response.Body.String()
	if strings.Count(body, `data-kind="set"`) != 1 || strings.Contains(body, `data-kind="rest"`) {
		t.Fatalf("playback did not render only the current slide")
	}
	if !strings.Contains(body, "3 kg") {
		t.Fatal("playback did not contain configured weight")
	}
	if !strings.Contains(body, `name="step" value="1"`) || !strings.Contains(body, `name="variant"`) {
		t.Fatal("playback did not contain server-driven progression state")
	}
	settings, err := application.store.playbackSettings()
	if err != nil || settings.RestSeconds != 30 || settings.ImagePreference != imageRandom {
		t.Fatalf("playback settings = (%#v, %v)", settings, err)
	}
}

func TestPlaybackAdvancesWithoutJavaScript(t *testing.T) {
	form := url.Values{
		"step": {"1"}, "rest_seconds": {"60"}, "image_preference": {imageMale},
		"weight":  {"2.5", "2.5", "2.5", "2.5", ""},
		"variant": {imageMale, imageMale, imageMale, imageMale, imageMale},
	}
	request := httptest.NewRequest(http.MethodPost, "/workouts/1/play", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	newTestApp(t).Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	body := response.Body.String()
	if !strings.Contains(body, `<section class="playback-slide" data-kind="rest"><p class="playback-progress">1 of 10 sets complete`) ||
		!strings.Contains(body, `role="timer"`) || strings.Contains(body, `aria-live=`) ||
		!strings.Contains(body, `<button type="submit">Next</button>`) {
		t.Fatal("server-driven playback did not render the accessible rest screen")
	}
}

func TestPlaybackCompletionSavesUnchangedWeights(t *testing.T) {
	application := newTestApp(t)
	form := url.Values{
		"step": {"19"}, "rest_seconds": {"60"}, "image_preference": {imageMale}, "save_weights": {"true"},
		"weight":  {"2.5", "2.5", "2.5", "2.5", ""},
		"variant": {imageMale, imageMale, imageMale, imageMale, imageMale},
	}
	request := httptest.NewRequest(http.MethodPost, "/workouts/1/play", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	application.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Working weights saved.") {
		t.Fatalf("completion response = (%d, %q)", response.Code, response.Body.String())
	}
}

func TestPlaybackCompletionWithoutJavaScriptSavesWeights(t *testing.T) {
	application := newTestApp(t)
	form := url.Values{
		"step": {"19"}, "rest_seconds": {"60"}, "image_preference": {imageMale}, "save_weights": {"true"},
		"weight":  {"8", "7", "6", "5", ""},
		"variant": {imageMale, imageMale, imageMale, imageMale, imageMale},
	}
	request := httptest.NewRequest(http.MethodPost, "/workouts/1/play", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	application.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Working weights saved.") {
		t.Fatalf("completion response = (%d, %q)", response.Code, response.Body.String())
	}
	workout, err := application.store.workout(1)
	if err != nil || workout.Exercises[0].Weight == nil || *workout.Exercises[0].Weight != 8 {
		t.Fatalf("saved workout = (%#v, %v)", workout, err)
	}
}

func TestPlaybackAlternatingImagesUseCookieState(t *testing.T) {
	form := url.Values{
		"rest_seconds": {"60"}, "image_preference": {imageAlternating},
		"weight": {"2.5", "2.5", "2.5", "2.5", ""},
	}
	request := httptest.NewRequest(http.MethodPost, "/workouts/1/play", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.AddCookie(&http.Cookie{Name: playbackAlternatingCookie, Value: imageFemale})
	response := httptest.NewRecorder()
	newTestApp(t).Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "goblet-squat-female.webp") {
		t.Fatalf("alternating response = (%d, %q)", response.Code, response.Body.String())
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != playbackAlternatingCookie || cookies[0].Value != imageMale {
		t.Fatalf("alternating cookies = %#v", cookies)
	}
}

func TestPlaybackRejectsInvalidSetup(t *testing.T) {
	form := url.Values{"rest_seconds": {"-1"}, "image_preference": {imageMale}, "weight": {"1"}}
	request := httptest.NewRequest(http.MethodPost, "/workouts/1/play", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	newTestApp(t).Handler().ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}
