package app

import (
	"crypto/rand"
	"fmt"
	"io"
)

const (
	imageMale        = "male"
	imageFemale      = "female"
	imageRandom      = "random"
	imageAlternating = "alternating"
)

func validImagePreference(preference string) bool {
	switch preference {
	case imageMale, imageFemale, imageRandom, imageAlternating:
		return true
	default:
		return false
	}
}

func selectImageVariant(preference, alternatingNext string) (variant, next string, err error) {
	switch preference {
	case imageMale, imageFemale:
		return preference, alternatingNext, nil
	case imageAlternating:
		if alternatingNext != imageFemale {
			return imageMale, imageFemale, nil
		}
		return imageFemale, imageMale, nil
	case imageRandom:
		return randomImageVariant(rand.Reader)
	default:
		return "", alternatingNext, fmt.Errorf("invalid image preference %q", preference)
	}
}

func randomImageVariant(source io.Reader) (variant, next string, err error) {
	var value [1]byte
	if _, err := io.ReadFull(source, value[:]); err != nil {
		return "", "", fmt.Errorf("choose random image variant: %w", err)
	}
	if value[0]&1 == 0 {
		return imageMale, "", nil
	}
	return imageFemale, "", nil
}
