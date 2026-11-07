package main

import "testing"

func TestEnvOrDefault(t *testing.T) {
	t.Setenv("SIMPLE_WORKOUT_TEST_VALUE", "configured")
	if got := envOrDefault("SIMPLE_WORKOUT_TEST_VALUE", "fallback"); got != "configured" {
		t.Fatalf("envOrDefault() = %q, want configured", got)
	}
}

func TestEnvOrDefaultFallback(t *testing.T) {
	if got := envOrDefault("SIMPLE_WORKOUT_MISSING_VALUE", "fallback"); got != "fallback" {
		t.Fatalf("envOrDefault() = %q, want fallback", got)
	}
}
