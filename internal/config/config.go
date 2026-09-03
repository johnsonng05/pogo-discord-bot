package config

import (
	"context"
	"errors"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
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
	appEnv := os.Getenv("APP_ENV")

	// Local development only
	if appEnv == "development" {
		cfg := &Config{
			DiscordToken:          os.Getenv("DISCORD_TOKEN"),
			AnnouncementChannelID: os.Getenv("ANNOUNCEMENT_CHANNEL_ID"),
			RedisURL:              os.Getenv("REDIS_URL"),
		}

		if cfg.DiscordToken == "" {
			return nil, errors.New("DISCORD_TOKEN is required")
		}

		if cfg.RedisURL == "" {
			return nil, errors.New("REDIS_URL is required")
		}
		return cfg, nil
	}

	// Production — pull from AWS Parameter Store.

	cfg := &Config{}

	ctx := context.Background()
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, errors.New("failed to load AWS SDK config: " + err.Error())
	}

	ssmClient := ssm.NewFromConfig(awsCfg)
	token, err := fetchAWSParameter(ctx, ssmClient, "/prod/discord/bot_token")
	if err != nil {
		return nil, errors.New("failed to retrieve bot token: " + err.Error())
	}
	cfg.DiscordToken = token

	channelID, err := fetchAWSParameter(ctx, ssmClient, "/prod/discord/announcement_channel_id")
	if err != nil {
		return nil, errors.New("failed to retrieve channel ID: " + err.Error())
	}
	cfg.AnnouncementChannelID = channelID

	redisURL, err := fetchAWSParameter(ctx, ssmClient, "/prod/discord/redis_url")
	if err != nil {
		return nil, errors.New("failed to retrieve Redis URL: " + err.Error())
	}
	cfg.RedisURL = redisURL

	return cfg, nil
}

func fetchAWSParameter(ctx context.Context, client *ssm.Client, name string) (string, error) {
	input := &ssm.GetParameterInput{
		Name:           &name,
		WithDecryption: aws.Bool(true),
	}
	result, err := client.GetParameter(ctx, input)
	if err != nil {
		return "", err
	}
	return *result.Parameter.Value, nil
}
