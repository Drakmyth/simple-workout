package app

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestRestSound(t *testing.T) {
	app := newTestApp(t)
	settings, err := app.store.playbackSettings()
	if err != nil || settings.RestSound {
		t.Fatalf("default settings = %+v, %v", settings, err)
	}
	for _, enabled := range []bool{true, false} {
		for _, step := range []string{"0", "1", "19", "20"} {
			form := url.Values{"step": {step}, "rest_seconds": {"1"}, "image_preference": {imageMale}, "weight": {"1", "2", "3", "4", "5"}}
			if enabled {
				form.Set("rest_sound", "true")
			}
			request := httptest.NewRequest(http.MethodPost, "/workouts/1/play", strings.NewReader(form.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			response := httptest.NewRecorder()
			app.Handler().ServeHTTP(response, request)
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d: %s", response.Code, response.Body.String())
			}
			body := response.Body.String()
			if strings.Contains(body, "data-rest-sound") != (enabled && step == "1") {
				t.Fatal("incorrect audio markup")
			}
			if strings.Contains(body, `name="rest_sound"`) != (enabled && step != "20") {
				t.Fatal("sound preference not carried forward")
			}
		}
		settings, err = app.store.playbackSettings()
		if err != nil || settings.RestSound != enabled {
			t.Fatalf("saved settings = %+v, %v", settings, err)
		}
		response := httptest.NewRecorder()
		app.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/workouts/1/start", nil))
		checked := `name="rest_sound" type="checkbox" value="true" checked`
		if strings.Contains(response.Body.String(), checked) != enabled {
			t.Fatal("setup did not remember sound preference")
		}
	}
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/static/rest-chime.wav", nil))
	if response.Code != http.StatusOK || !strings.HasPrefix(response.Body.String(), "RIFF") {
		t.Fatal("chime asset not served")
	}
}
