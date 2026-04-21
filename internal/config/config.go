// ABOUTME: Loads MCP server configuration from environment variables.
// ABOUTME: INTERVALS_API_KEY (required), INTERVALS_ATHLETE_ID and INTERVALS_BASE_URL (optional).

package config

import (
	"errors"
	"os"
	"strings"
)

const (
	EnvAPIKey    = "INTERVALS_API_KEY"
	EnvAthleteID = "INTERVALS_ATHLETE_ID"
	EnvBaseURL   = "INTERVALS_BASE_URL"

	DefaultAthleteID = "0"
	DefaultBaseURL   = "https://intervals.icu/api/v1"
)

type Config struct {
	APIKey    string
	AthleteID string
	BaseURL   string
}

var ErrMissingAPIKey = errors.New("INTERVALS_API_KEY is required")

func Load() (Config, error) {
	key := strings.TrimSpace(os.Getenv(EnvAPIKey))
	if key == "" {
		return Config{}, ErrMissingAPIKey
	}
	athlete := strings.TrimSpace(os.Getenv(EnvAthleteID))
	if athlete == "" {
		athlete = DefaultAthleteID
	}
	baseURL := strings.TrimSpace(os.Getenv(EnvBaseURL))
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return Config{
		APIKey:    key,
		AthleteID: athlete,
		BaseURL:   baseURL,
	}, nil
}
