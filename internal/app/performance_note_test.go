package app

import (
	"bytes"
	"html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestPerformanceNotesAndSummary(t *testing.T) {
	app := newTestApp(t)
	form := url.Values{"rest_seconds": {"60"}, "image_preference": {imageMale},
		"weight": {"1", "2", "3", "4", "0"}, "save_weights": {"true"}}
	post := func(step string) string {
		t.Helper()
		form.Set("step", step)
		request := httptest.NewRequest(http.MethodPost, "/workouts/1/play", strings.NewReader(form.Encode()))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		response := httptest.NewRecorder()
		app.Handler().ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("status %d: %s", response.Code, response.Body.String())
		}
		return response.Body.String()
	}
	if body := post("1"); strings.Contains(body, "<textarea") {
		t.Fatal("note shown between sets")
	}
	body := post("3")
	if !strings.Contains(body, `name="note_0" form="playback-action"`) {
		t.Fatal("missing first exercise note")
	}
	form.Set("note_0", "Tough <script> & steady\nGood form")
	body = post("4")
	if !strings.Contains(body, `name="note_0" value="Tough &lt;script&gt; &amp; steady`) {
		t.Fatal("note not safely carried forward")
	}
	body = post("19")
	for _, want := range []string{"Workout complete", "Working weights saved.", "View summary", `name="note_4" form="playback-action"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("completion missing %q", want)
		}
	}
	// Opening the summary must not save weights again.
	if _, err := app.store.db.Exec("UPDATE workout_exercises SET weight = 99 WHERE workout_id = 1"); err != nil {
		t.Fatal(err)
	}
	form.Set("note_4", "Finished strong")
	body = post("20")
	for _, want := range []string{"Workout summary", "2 sets × 10 reps", "0 kg", "Tough &lt;script&gt; &amp; steady", "Finished strong"} {
		if !strings.Contains(body, want) {
			t.Fatalf("summary missing %q: %s", want, body)
		}
	}
	if strings.Contains(body, "<textarea") || strings.Contains(body, "data-rest-timer") || strings.Contains(body, "data-keep-screen-on") {
		t.Fatal("summary is not read-only")
	}
	workout, err := app.store.playbackWorkout(1)
	if err != nil || *workout.Exercises[0].Weight != 99 {
		t.Fatal("summary resaved weights")
	}
}

func TestPerformanceNotesRemainPerOccurrence(t *testing.T) {
	exercises := playbackExerciseViews([]PlaybackExercise{
		{Name: "Curl", Sets: 1}, {Name: "Curl", Sets: 1},
	})
	if exercises[0].Index == exercises[1].Index {
		t.Fatal("duplicate exercises share note identity")
	}
	exercises[0].Note = "First occurrence"
	exercises[1].Note = "Second occurrence"
	var body bytes.Buffer
	app := newTestApp(t)
	if err := app.templates.ExecuteTemplate(&body, "playback-summary", playbackPage{Exercises: exercises}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"First occurrence", "Second occurrence", "Bodyweight"} {
		if !strings.Contains(body.String(), want) {
			t.Fatalf("duplicate summary missing %q", want)
		}
	}
}

func TestCopySummaryText(t *testing.T) {
	app := newTestApp(t)
	exercises := playbackExerciseViews([]PlaybackExercise{
		{Name: "Curl", Sets: 2, Reps: 10},
		{Name: "Squat", Sets: 3, Reps: 8},
	})
	exercises[1].Weight = "0"
	exercises[1].Note = "Tough <but safe>\nGood finish"
	var body bytes.Buffer
	if err := app.templates.ExecuteTemplate(&body, "playback-summary", playbackPage{
		Workout: PlaybackWorkout{Name: "Full body"}, Exercises: exercises, Unit: "lb",
	}); err != nil {
		t.Fatal(err)
	}
	_, text, ok := strings.Cut(body.String(), `<pre data-summary-text hidden>`)
	if !ok {
		t.Fatal("missing clipboard text")
	}
	text, _, _ = strings.Cut(text, "</pre>")
	want := "Workout: Full body\n\nCurl\n2 sets × 10 reps · Bodyweight\n\nSquat\n3 sets × 8 reps · 0 lb\nNotes: Tough <but safe>\nGood finish"
	if got := html.UnescapeString(text); got != want {
		t.Fatalf("clipboard text = %q, want %q", got, want)
	}
}

func TestRejectOversizedPerformanceNote(t *testing.T) {
	app := newTestApp(t)
	form := url.Values{"rest_seconds": {"60"}, "image_preference": {imageMale},
		"weight": {"1", "2", "3", "4", "5"}, "note_0": {strings.Repeat("x", 4001)}}
	request := httptest.NewRequest(http.MethodPost, "/workouts/1/play", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", response.Code)
	}
}
