package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type entry struct {
	Exercise       string `json:"exercise"`
	Variant        string `json:"variant"`
	ModelReference string `json:"modelReference"`
	PromptRecipe   string `json:"promptRecipe"`
	SourceSHA256   string `json:"sourceSha256"`
	Output         string `json:"output"`
	ReviewStatus   string `json:"reviewStatus"`
}

func main() {
	sourceDir := flag.String("source", "reference/exercises", "directory containing reviewed source PNGs")
	requirementsPath := flag.String("requirements", "assets/exercise-images/requirements.json", "exercise requirements file")
	outputPath := flag.String("output", "assets/exercise-images/manifest.json", "manifest output path")
	flag.Parse()

	if err := run(*sourceDir, *requirementsPath, *outputPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(sourceDir, requirementsPath, outputPath string) error {
	requirementsData, err := os.ReadFile(requirementsPath)
	if err != nil {
		return fmt.Errorf("read requirements: %w", err)
	}
	var requirements map[string]json.RawMessage
	if err := json.Unmarshal(requirementsData, &requirements); err != nil {
		return fmt.Errorf("parse requirements: %w", err)
	}

	var entries []entry
	for exercise := range requirements {
		for _, variant := range []string{"female", "male"} {
			name := exercise + "-" + variant
			path := filepath.Join(sourceDir, name+".png")
			hash, err := validateSource(path)
			if err != nil {
				return err
			}
			entries = append(entries, entry{
				Exercise:       exercise,
				Variant:        variant,
				ModelReference: "assets/exercise-images/models/" + variant + ".png",
				PromptRecipe:   "assets/exercise-images/requirements.json#" + exercise,
				SourceSHA256:   hash,
				Output:         "internal/app/static/images/exercises/" + name + ".webp",
				ReviewStatus:   "approved",
			})
		}
	}
	slices.SortFunc(entries, func(a, b entry) int {
		return strings.Compare(a.Exercise+"/"+a.Variant, b.Exercise+"/"+b.Variant)
	})

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("encode manifest: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(outputPath, data, 0o644); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}
	return nil
}

func validateSource(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open source %s: %w", path, err)
	}
	defer file.Close()

	config, format, err := image.DecodeConfig(file)
	if err != nil {
		return "", fmt.Errorf("decode source %s: %w", path, err)
	}
	if format != "png" || config.Width != 1536 || config.Height != 1024 {
		return "", fmt.Errorf("source %s must be a 1536x1024 PNG", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read source %s: %w", path, err)
	}
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum), nil
}
