package commands

import (
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const layout = "2006-01-02T15:04:05.000"

// Events handles /pogo-events.
// Fetch the event timeline, pick what is live / upcoming, and render a Discord embed.
func (h *Handler) CurrentEvents(s *discordgo.Session, i *discordgo.InteractionCreate) {
	events, err := h.API.FetchEvents()
	if err != nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Error fetching events",
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Error fetching events",
					Description: "Error fetching events",
					Color:       0x0099ff,
				},
			},
		})
		return
	}

	var embeds []*discordgo.MessageEmbed
	for _, event := range events {
		start := strings.TrimSuffix(event.Start, "Z")
		end := strings.TrimSuffix(event.End, "Z")
		startTime, err := time.Parse(layout, start)
		if err != nil {
			continue
		}
		endTime, err := time.Parse(layout, end)
		if err != nil {
			continue
		}

		currentTime := time.Now()
		if currentTime.Before(startTime) || currentTime.After(endTime) {
			continue
		}

		if event.Heading == "GO Battle League" || event.Heading == "GO Pass" || event.Heading == "Season" {
			continue
		}

		embeds = append(embeds, &discordgo.MessageEmbed{
			Title:       event.Name,
			URL:         event.Link,
			Description: startTime.Format("Jan 2 15:04") + " – " + endTime.Format("Jan 2 15:04"),
			Color:       0x0099ff,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: event.Image,
			},
		})
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

func (h *Handler) UpcomingEvents(s *discordgo.Session, i *discordgo.InteractionCreate) {
	events, err := h.API.FetchEvents()
	if err != nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Error fetching events",
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Error fetching upcoming events",
					Description: "There was an error fetching the upcoming events",
					Color:       0x0099ff,
				},
			},
		})
	}

	var embeds []*discordgo.MessageEmbed
	for _, event := range events {
		event.Start = strings.TrimSuffix(event.Start, "Z")
		event.End = strings.TrimSuffix(event.End, "Z")

		startTime, err := time.Parse(layout, event.Start)
		if err != nil {
			continue
		}

		endTime, err := time.Parse(layout, event.End)
		if err != nil {
			continue
		}

		if event.Heading == "GO Battle League" || event.Heading == "GO Pass" || event.Heading == "Season" {
			continue
		}

		currentTime := time.Now()
		if currentTime.Before(startTime) {
			embeds = append(embeds, &discordgo.MessageEmbed{
				Title:       event.Name,
				URL:         event.Link,
				Description: startTime.Format("Jan 2 15:04") + " – " + endTime.Format("Jan 2 15:04"),
				Color:       0x0099ff,
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: event.Image,
				},
			})
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
