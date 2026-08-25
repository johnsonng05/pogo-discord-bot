package config

import "os"

// Config holds process-level settings loaded from the environment.
// See README.md → "Environment Configuration Matrix".
type Config struct {
	DiscordToken          string
	AnnouncementChannelID string
}

// Load reads DISCORD_TOKEN and ANNOUNCEMENT_CHANNEL_ID.
func Load() (*Config, error) {
	cfg := &Config{
		DiscordToken:          os.Getenv("DISCORD_TOKEN"),
		AnnouncementChannelID: os.Getenv("ANNOUNCEMENT_CHANNEL_ID"),
	}

	if cfg.DiscordToken == "" {
		return nil, errors.New("DISCORD_TOKEN is required")
	}

	return cfg, nil
}
