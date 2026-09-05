package commands

import (
	"fmt"
	"strings"

	"pogo-discord-bot/internal/models"

	"github.com/bwmarrin/discordgo"
)

// Raids handles /pogo-raids.
// Fetch current raid bosses and group them by tier in an embed.
func (h *Handler) Raids(s *discordgo.Session, i *discordgo.InteractionCreate) {

	tierColors := map[string]int{
		"Super Mega Raids": 0xFF00FF, // Purple
		"Mega Raids":       0xFF00FF, // Purple
		"6-Star Raids":     0xFF0000, // Red
		"5-Star Raids":     0xFF0000, // Red
		"4-Star Raids":     0xFFFF00, // Yellow
		"3-Star Raids":     0xFFFF00, // Yellow
		"2-Star Raids":     0x00FF00, // Green
		"1-Star Raids":     0x00FF00, // Green
	}

	tierOrder := []string{
		"Super Mega Raids",
		"Mega Raids",
		"6-Star Raids",
		"5-Star Raids",
		"4-Star Raids",
		"3-Star Raids",
		"2-Star Raids",
		"1-Star Raids",
	}

	raids, err := h.API.FetchRaidBosses()
	if err != nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Error fetching raids",
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Error fetching raids",
					Description: "Sorry, there was an error fetching the raids...",
					Color:       0x0099ff,
				},
			},
		})
		return
	}

	byTier := map[string][]models.RaidBoss{}
	for _, raid := range raids {
		byTier[raid.Tier] = append(byTier[raid.Tier], raid)
	}

	var embeds []*discordgo.MessageEmbed

	for _, tier := range tierOrder {
		bosses, ok := byTier[tier]
		if !ok || len(bosses) == 0 {
			continue
		}

		embed := &discordgo.MessageEmbed{
			Title: tier,
			Color: tierColors[tier],
		}

		if bosses[0].Image != "" {
			embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: bosses[0].Image}
		}

		for _, pokemon := range bosses {
			typeNames := make([]string, 0, len(pokemon.Types))
			for _, t := range pokemon.Types {
				typeNames = append(typeNames, t.Name)
			}

			weatherNames := make([]string, 0, len(pokemon.BoostedWeather))
			for _, w := range pokemon.BoostedWeather {
				weatherNames = append(weatherNames, w.Name)
			}

			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name: pokemon.Name,
				Value: fmt.Sprintf(
					"Type: %s\nPossible Shiny: %t\nBoosted Weather: %s\nCP: %d–%d (boosted %d–%d)",
					strings.Join(typeNames, ", "),
					pokemon.CanBeShiny,
					strings.Join(weatherNames, ", "),
					pokemon.CombatPower.Normal.Min,
					pokemon.CombatPower.Normal.Max,
					pokemon.CombatPower.Boosted.Min,
					pokemon.CombatPower.Boosted.Max,
				),
				Inline: true,
			})
		}

		embeds = append(embeds, embed)
	}

	if len(embeds) == 0 {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "No raids found",
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "No raids currently",
					Description: "Sorry, no raids were found at the moment...",
					Color:       0x0099ff,
				},
			},
		})
		return
	} else {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Here are the current raids!",
			Embeds:  embeds,
		})
	}
}
