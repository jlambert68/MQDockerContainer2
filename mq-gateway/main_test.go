package main

import "testing"

func TestGetenv(t *testing.T) {
	t.Setenv("MAIN_TEST_ENV", "value")
	if got := getenv("MAIN_TEST_ENV", "default"); got != "value" {
		t.Fatalf("expected env value, got %q", got)
	}
	if got := getenv("MAIN_TEST_ENV_MISSING", "default"); got != "default" {
		t.Fatalf("expected default value, got %q", got)
	}
}
