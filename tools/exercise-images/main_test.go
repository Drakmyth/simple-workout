package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateSource(t *testing.T) {
	path := writePNG(t, 1536, 1024)
	hash, err := validateSource(path)
	if err != nil {
		t.Fatalf("validateSource() error = %v", err)
	}
	if len(hash) != 64 {
		t.Fatalf("hash length = %d, want 64", len(hash))
	}
}

func TestValidateSourceRejectsUnexpectedDimensions(t *testing.T) {
	path := writePNG(t, 10, 10)
	_, err := validateSource(path)
	if err == nil || !strings.Contains(err.Error(), "must be a 1536x1024 PNG") {
		t.Fatalf("validateSource() error = %v", err)
	}
}

func writePNG(t *testing.T, width, height int) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "image.png")
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create PNG: %v", err)
	}
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	canvas.Set(0, 0, color.White)
	if err := png.Encode(file, canvas); err != nil {
		file.Close()
		t.Fatalf("encode PNG: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close PNG: %v", err)
	}
	return path
}
