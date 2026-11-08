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
	if strings.Count(body, `data-kind="set"`) != 10 || strings.Count(body, `data-kind="rest"`) != 9 {
		t.Fatalf("playback slide counts were incorrect")
	}
	if !strings.Contains(body, `data-image-preference="random"`) || !strings.Contains(body, "3 kg") {
		t.Fatal("playback did not contain configured image preference and weight")
	}
	settings, err := application.store.playbackSettings()
	if err != nil || settings.RestSeconds != 30 || settings.ImagePreference != imageRandom {
		t.Fatalf("playback settings = (%#v, %v)", settings, err)
	}
}

func TestPlaybackUsesOneExerciseIndexAcrossSets(t *testing.T) {
	exercises := []playbackExerciseView{
		{PlaybackExercise: PlaybackExercise{Name: "First", Sets: 2}},
		{PlaybackExercise: PlaybackExercise{Name: "Second", Sets: 1}},
	}
	slides := buildPlaybackSlides(exercises, 60)
	var indexes []int
	for _, slide := range slides {
		if slide.Kind == "set" {
			indexes = append(indexes, slide.ExerciseIndex)
		}
	}
	if len(indexes) != 3 || indexes[0] != 0 || indexes[1] != 0 || indexes[2] != 1 {
		t.Fatalf("exercise indexes = %v, want [0 0 1]", indexes)
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
