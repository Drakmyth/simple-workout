package app

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestKeepScreenOn(t *testing.T) {
	app := newTestApp(t)
	settings, err := app.store.playbackSettings()
	if err != nil || !settings.KeepScreenOn {
		t.Fatalf("default settings = %+v, %v", settings, err)
	}
	for _, enabled := range []bool{false, true} {
		for _, step := range []string{"0", "1", "19"} {
			form := url.Values{"step": {step}, "rest_seconds": {"60"}, "image_preference": {imageMale}, "weight": {"1", "2", "3", "4", "5"}}
			if enabled {
				form.Set("keep_screen_on", "true")
			}
			request := httptest.NewRequest(http.MethodPost, "/workouts/1/play", strings.NewReader(form.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			response := httptest.NewRecorder()
			app.Handler().ServeHTTP(response, request)
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d: %s", response.Code, response.Body.String())
			}
			body := response.Body.String()
			want := enabled && step != "19"
			if strings.Contains(body, "data-keep-screen-on") != want || strings.Contains(body, `name="keep_screen_on"`) != enabled {
				t.Fatalf("incorrect wake lock markup: %s", body)
			}
		}
		settings, err = app.store.playbackSettings()
		if err != nil || settings.KeepScreenOn != enabled {
			t.Fatalf("saved settings = %+v, %v", settings, err)
		}
	}
}
