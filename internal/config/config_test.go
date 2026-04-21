// ABOUTME: Tests for config loader (env var parsing and defaults).
// ABOUTME: Covers missing API key, default athlete id, and trim behavior.

package config

import (
	"errors"
	"testing"
)

func TestLoad_MissingAPIKey(t *testing.T) {
	t.Setenv(EnvAPIKey, "")
	t.Setenv(EnvAthleteID, "")
	t.Setenv(EnvBaseURL, "")

	_, err := Load()
	if !errors.Is(err, ErrMissingAPIKey) {
		t.Fatalf("expected ErrMissingAPIKey, got %v", err)
	}
}

func TestLoad_DefaultsAthleteID(t *testing.T) {
	t.Setenv(EnvAPIKey, "secret")
	t.Setenv(EnvAthleteID, "")
	t.Setenv(EnvBaseURL, "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.APIKey != "secret" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "secret")
	}
	if cfg.AthleteID != DefaultAthleteID {
		t.Errorf("AthleteID = %q, want %q", cfg.AthleteID, DefaultAthleteID)
	}
	if cfg.BaseURL != DefaultBaseURL {
		t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, DefaultBaseURL)
	}
}

func TestLoad_ExplicitAthleteIDAndTrim(t *testing.T) {
	t.Setenv(EnvAPIKey, "  secret  ")
	t.Setenv(EnvAthleteID, "  12345  ")
	t.Setenv(EnvBaseURL, "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.APIKey != "secret" {
		t.Errorf("APIKey = %q, want trimmed", cfg.APIKey)
	}
	if cfg.AthleteID != "12345" {
		t.Errorf("AthleteID = %q, want %q", cfg.AthleteID, "12345")
	}
}

func TestLoad_BaseURLOverride(t *testing.T) {
	t.Setenv(EnvAPIKey, "secret")
	t.Setenv(EnvAthleteID, "")
	t.Setenv(EnvBaseURL, "  https://staging.intervals.icu/api/v1  ")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.BaseURL != "https://staging.intervals.icu/api/v1" {
		t.Errorf("BaseURL = %q, want override (trimmed)", cfg.BaseURL)
	}
}

func TestLoad_BaseURLEmptyFallsBackToDefault(t *testing.T) {
	t.Setenv(EnvAPIKey, "secret")
	t.Setenv(EnvAthleteID, "")
	t.Setenv(EnvBaseURL, "   ")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.BaseURL != DefaultBaseURL {
		t.Errorf("BaseURL = %q, want default %q", cfg.BaseURL, DefaultBaseURL)
	}
}
