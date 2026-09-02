package cache

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestNewPing(t *testing.T) {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		t.Skip("REDIS_URL not set")
	}
	c, err := New(url)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
}

func TestAnnouncementChannels(t *testing.T) {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		t.Skip("REDIS_URL not set")
	}

	c, err := New(url)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	ctx := context.Background()
	guildID := "test-guild-1"
	channelID := "test-channel-99"

	if err := c.SetAnnouncementChannel(ctx, guildID, channelID); err != nil {
		t.Fatal(err)
	}

	got, err := c.GetAnnouncementChannel(ctx, guildID)
	if err != nil {
		t.Fatal(err)
	}
	if got != channelID {
		t.Fatalf("got %q, want %q", got, channelID)
	}

	all, err := c.ListAnnouncementChannels(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if all[guildID] != channelID {
		t.Fatalf("HGETALL: got %v", all)
	}
}

func TestSetCachedTTL(t *testing.T) {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		t.Skip("REDIS_URL not set")
	}

	c, err := New(url)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	ctx := context.Background()
	key := "pogo:test:ttl"
	defer c.Delete(ctx, key)

	const ttl = 30 * time.Second
	payload := []byte(`{"ok":true}`)

	if err := c.SetCached(ctx, key, payload, ttl); err != nil {
		t.Fatal(err)
	}

	remaining, err := c.TTL(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if remaining <= 0 || remaining > ttl {
		t.Fatalf("TTL = %v, want > 0 and <= %v", remaining, ttl)
	}

	got, err := c.GetCached(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(payload) {
		t.Fatalf("got %q, want %q", got, payload)
	}
}

func TestCachedEntryExpires(t *testing.T) {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		t.Skip("REDIS_URL not set")
	}

	c, err := New(url)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	ctx := context.Background()
	key := "pogo:test:expire"
	defer c.Delete(ctx, key)

	if err := c.SetCached(ctx, key, []byte("gone"), time.Second); err != nil {
		t.Fatal(err)
	}

	time.Sleep(1100 * time.Millisecond)

	_, err = c.GetCached(ctx, key)
	if !errors.Is(err, redis.Nil) {
		t.Fatalf("expected redis.Nil after expiry, got %v", err)
	}
}
