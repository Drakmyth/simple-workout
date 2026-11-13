package app

import (
	"bytes"
	"strings"
	"testing"
)

func TestRestPreview(t *testing.T) {
	for _, weight := range []string{"", "0", "22.5"} {
		exercises := []playbackExerciseView{
			{PlaybackExercise: PlaybackExercise{Name: "Curl", Sets: 2, Reps: 10}, Weight: weight},
			{PlaybackExercise: PlaybackExercise{Name: "Squat", Sets: 1, Reps: 8}, Weight: weight},
		}
		slides := buildPlaybackSlides(exercises, 30, "lb")
		for _, test := range []struct {
			index, set int
			name       string
		}{
			{1, 2, "Curl"}, {3, 1, "Squat"},
		} {
			slide := slides[test.index]
			if slide.Exercise.Name != test.name || slide.SetNumber != test.set || slide.Exercise.Weight != weight {
				t.Fatalf("unexpected preview: %+v", slide)
			}
		}
	}
}

func TestRestPreviewRendering(t *testing.T) {
	app := newTestApp(t)
	for _, test := range []struct{ weight, want string }{
		{"", "Bodyweight"}, {"0", "0 lb"}, {"22.5", "22.5 lb"},
	} {
		slide := playbackSlide{Kind: "rest", Unit: "lb", SetNumber: 2, RestSeconds: 30,
			Exercise: playbackExerciseView{PlaybackExercise: PlaybackExercise{Name: "Curl", Sets: 3, Reps: 10}, Weight: test.weight}}
		var body bytes.Buffer
		if err := app.templates.ExecuteTemplate(&body, "playback-slide", slide); err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{"Up next: Curl", "Set 2 of 3", "10 reps", test.want} {
			if !strings.Contains(body.String(), want) {
				t.Fatalf("missing %q in %s", want, body.String())
			}
		}
	}
}
