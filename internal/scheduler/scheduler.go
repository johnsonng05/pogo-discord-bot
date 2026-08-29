package scheduler

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"pogo-bot/internal/api"
	"pogo-bot/internal/models"

	"github.com/bwmarrin/discordgo"
)

// TODO: Add raids to the scheduler
// TODO: Refactor shared event helpers into one package
// TODO: Load per-guild channels from Redis at post time

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
	// Store *store.Store // TODO: load per-guild channels from Redis at post time
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

// Run loops until Stop is called.
func (s *Scheduler) Run() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	var lastPosted string
	location := eventLocation()

	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:

			now := time.Now().In(location)
			if now.Hour() != s.TargetHour {
				continue
			}
			today := now.Format("2006-01-02")
			if today == lastPosted {
				continue
			}

			if s.ChannelID == "" {
				log.Println("scheduler: ANNOUNCEMENT_CHANNEL_ID not set")
				continue
			}

			s.postDaily()
			lastPosted = today
		}
	}
}

// Stop signals Run to exit. Safe to call from Bot.Close
// Use a sync.Once to ensure the channel is closed only once (double-close / panic prevention)
func (s *Scheduler) Stop() {
	s.stopOnce.Do(func() {
		close(s.stop)
	})
}

// postDaily is a helper you can call from Run once the 08:00 condition matches.
func (s *Scheduler) postDaily() {
	events, err := s.API.FetchEvents()
	currentTime := time.Now().In(eventLocation())

	fields := make([]*discordgo.MessageEmbedField, 0)
	embed := &discordgo.MessageEmbed{
		Title:       "Events",
		Description: "Here are the events for today",
		Color:       0x0099ff,
		Fields:      fields,
	}

	if err != nil {
		log.Println("Error fetching events", err)
		return
	}

	for _, event := range events {
		if isExcludedEvent(&event) {
			continue
		}

		startTime, endTime, err := parseEventTime(event.Start, event.End)
		if err != nil {
			continue
		}

		if currentTime.Before(startTime) || currentTime.After(endTime) {
			continue
		}

		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  event.Name,
			Value: fmt.Sprintf("%s – %s PT", startTime.Format("January 2, 2006 3:04 PM"), endTime.Format("January 2, 2006 3:04 PM")),
		})

		embed.Fields = fields
	}

	s.Session.ChannelMessageSendComplex(s.ChannelID, &discordgo.MessageSend{
		Content: "☀️ **Daily Pokémon GO brief**",
		Embeds:  []*discordgo.MessageEmbed{embed},
	})
}

func isExcludedEvent(event *models.Event) bool {
	return event.Heading == "GO Battle League" ||
		event.Heading == "GO Pass" ||
		event.Heading == "Season"
}

func parseEventTime(start, end string) (time.Time, time.Time, error) {
	location := eventLocation()

	start = strings.TrimSuffix(start, "Z")
	end = strings.TrimSuffix(end, "Z")

	startTime, err := time.ParseInLocation(layout, start, location)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	endTime, err := time.ParseInLocation(layout, end, location)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return startTime, endTime, nil
}

func eventEmbed(event *models.Event, startTime, endTime time.Time) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: event.Name,
		URL:   event.Link,
		Description: fmt.Sprintf(
			"%s – %s PT",
			startTime.Format("January 2, 2006 3:04 PM"),
			endTime.Format("January 2, 2006 3:04 PM"),
		),
		Color: 0x0099ff,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: event.Image,
		},
	}
}

func eventLocation() *time.Location {
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return time.UTC
	}
	return location
}
