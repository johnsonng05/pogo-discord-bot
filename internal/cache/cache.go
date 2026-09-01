package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Redis key conventions:
//   announcement:channels  — HASH: guildID → channelID (daily posts)

const (
	announcementChannelsKey = "announcement:channels"
)

// Cache persists guild settings and API response payloads in Redis.
type Cache struct {
	client *redis.Client
}

func New(redisURL string) (*Cache, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	opts.PoolSize = 10
	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return &Cache{client: client}, nil
}

// Close releases the Redis connection pool.
func (c *Cache) Close() error {
	return c.client.Close()
}

// SetAnnouncementChannel maps a guild to the channel that should receive daily posts.
func (c *Cache) SetAnnouncementChannel(ctx context.Context, guildID, channelID string) error {
	return c.client.HSet(ctx, announcementChannelsKey, guildID, channelID).Err()
}

// GetAnnouncementChannel returns the configured channel for one guild.
func (c *Cache) GetAnnouncementChannel(ctx context.Context, guildID string) (string, error) {
	return c.client.HGet(ctx, announcementChannelsKey, guildID).Result()
}

// ListAnnouncementChannels returns guildID → channelID for every configured server.
func (c *Cache) ListAnnouncementChannels(ctx context.Context) (map[string]string, error) {
	return c.client.HGetAll(ctx, announcementChannelsKey).Result()
}
