package config

import (
	"errors"
	"os"
)

// Config holds process-level settings loaded from the environment.
// See README.md → "Environment Configuration Matrix".
type Config struct {
	DiscordToken          string
	AnnouncementChannelID string
	RedisURL              string
}

// Load reads DISCORD_TOKEN, ANNOUNCEMENT_CHANNEL_ID, and REDIS_URL.
func Load() (*Config, error) {
	cfg := &Config{
		DiscordToken:          os.Getenv("DISCORD_TOKEN"),
		AnnouncementChannelID: os.Getenv("ANNOUNCEMENT_CHANNEL_ID"),
		RedisURL:              os.Getenv("REDIS_URL"),
	}

	if cfg.DiscordToken == "" {
		return nil, errors.New("DISCORD_TOKEN is required")
	}

	return cfg, nil
}
