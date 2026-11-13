package app

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestKeepImageVisible(t *testing.T) {
	app := newTestApp(t)
	settings, err := app.store.playbackSettings()
	if err != nil || settings.KeepImageVisible {
		t.Fatalf("default settings = %+v, %v", settings, err)
	}
	for _, enabled := range []bool{true, false} {
		for _, step := range []string{"0", "1", "19", "20"} {
			form := url.Values{"step": {step}, "rest_seconds": {"60"}, "image_preference": {imageMale}, "weight": {"1", "2", "3", "4", "5"}}
			if enabled {
				form.Set("keep_image_visible", "true")
			}
			request := httptest.NewRequest(http.MethodPost, "/workouts/1/play", strings.NewReader(form.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			response := httptest.NewRecorder()
			app.Handler().ServeHTTP(response, request)
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d: %s", response.Code, response.Body.String())
			}
			body := response.Body.String()
			if strings.Contains(body, "data-keep-image-visible") != (enabled && step == "0") {
				t.Fatal("incorrect pinned-image markup")
			}
			if strings.Contains(body, `name="keep_image_visible"`) != (enabled && step != "20") {
				t.Fatal("preference not carried forward")
			}
			if step == "0" && !strings.Contains(body, `tabindex="0" aria-label="Exercise guidance"`) {
				t.Fatal("guidance cannot be scrolled by keyboard")
			}
		}
		settings, err = app.store.playbackSettings()
		if err != nil || settings.KeepImageVisible != enabled {
			t.Fatalf("saved settings = %+v, %v", settings, err)
		}
		response := httptest.NewRecorder()
		app.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/workouts/1/start", nil))
		checked := `name="keep_image_visible" type="checkbox" value="true" checked`
		if strings.Contains(response.Body.String(), checked) != enabled {
			t.Fatal("setup did not remember image preference")
		}
	}
}
