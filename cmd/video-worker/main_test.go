package main

import "testing"

func TestGetenvReturnsConfiguredValue(t *testing.T) {
	t.Setenv("TEST_CONFIG", "configured")

	if got := getenv("TEST_CONFIG", "fallback"); got != "configured" {
		t.Fatalf("expected configured value, got %q", got)
	}
}

func TestGetenvReturnsFallbackWhenUnset(t *testing.T) {
	t.Setenv("TEST_CONFIG", "")

	if got := getenv("TEST_CONFIG", "fallback"); got != "fallback" {
		t.Fatalf("expected fallback value, got %q", got)
	}
}
