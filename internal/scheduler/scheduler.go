package scheduler

import (
	"sync"

	"pogo-bot/internal/api"

	"github.com/bwmarrin/discordgo"
)

// TargetHour is local server time for the daily announcement (README: 08:00 AM).
const (
	layout   = "2006-01-02T15:04:05.000"
	timezone = "America/Los_Angeles"
)

// Scheduler is the autonomous background routine.
// It must not block the interactive slash-command pipeline.
type Scheduler struct {
	Session    *discordgo.Session
	ChannelID  string
	API        *api.Client
	stop       chan struct{}
	stopOnce   sync.Once
	TargetHour int
}

func New(session *discordgo.Session, channelID string, client *api.Client) *Scheduler {
	return &Scheduler{
		Session:    session,
		ChannelID:  channelID,
		API:        client,
		stop:       make(chan struct{}),
		TargetHour: 8,
	}
}
