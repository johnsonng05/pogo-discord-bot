package cache

import (
	"context"
	"os"
	"testing"
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
