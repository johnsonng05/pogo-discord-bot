package commands

import (
	"fmt"
	"strings"
	"time"

	"pogo-discord-bot/internal/models"

	"github.com/bwmarrin/discordgo"
)

const (
	layout   = "2006-01-02T15:04:05.000"
	timezone = "America/Los_Angeles"
)

// CurrentEvents handles /pogo-current-events.
// Fetch the event timeline, pick what is live, and render a Discord embed.
func (h *Handler) CurrentEvents(s *discordgo.Session, i *discordgo.InteractionCreate) {
	events, err := h.API.FetchEvents()
	if err != nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Error fetching events",
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Error fetching events",
					Description: "Sorry, there was an error fetching the events...",
					Color:       0x0099ff,
				},
			},
		})
		return
	}

	var embeds []*discordgo.MessageEmbed
	currentTime := time.Now().In(eventLocation())

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

		embeds = append(embeds, eventEmbed(&event, startTime, endTime))
	}

	if len(embeds) == 0 {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "No events found",
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "No events found",
					Description: "No live events right now",
					Color:       0x0099ff,
				},
			},
		})
		return
	}

	s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: "Here are the live events happening right now!",
		Embeds:  embeds,
	})
}

// UpcomingEvents handles /pogo-upcoming-events.
// Fetch the event timeline, pick what is upcoming, and render a Discord embed.
func (h *Handler) UpcomingEvents(s *discordgo.Session, i *discordgo.InteractionCreate) {
	events, err := h.API.FetchEvents()
	if err != nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Error fetching upcoming events",
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Error fetching upcoming events",
					Description: "Sorry, there was an error fetching the upcoming events...",
					Color:       0x0099ff,
				},
			},
		})
		return
	}

	var embeds []*discordgo.MessageEmbed
	currentTime := time.Now().In(eventLocation())

	for _, event := range events {
		if isExcludedEvent(&event) {
			continue
		}

		startTime, endTime, err := parseEventTime(event.Start, event.End)
		if err != nil {
			continue
		}

		if currentTime.Before(startTime) {
			embeds = append(embeds, eventEmbed(&event, startTime, endTime))
		}
	}

	if len(embeds) > 10 {
		embeds = embeds[:10]
	}

	s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: "Here are the upcoming events happening soon!",
		Embeds:  embeds,
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
