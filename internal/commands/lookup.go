package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Lookup handles /pokemon-lookup.
func (h *Handler) Lookup(s *discordgo.Session, i *discordgo.InteractionCreate) {
	targetPokemon := i.ApplicationCommandData().Options[0].StringValue()
	if targetPokemon == "" {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Please provide a valid Pokémon name.",
		})
		return
	}

	pokemonProfile, err := h.API.LookupPokemon(targetPokemon)
	if err != nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Failed to lookup Pokémon.",
		})
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("**%s** ", pokemonProfile.Stats.PokemonName),
		Description: fmt.Sprintf("**Form**\n%s", pokemonProfile.Stats.Form),
		Color:       0x0099ff,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "**Base Stats**", Value: fmt.Sprintf("Attack: %d\nDefense: %d\nStamina: %d", pokemonProfile.Stats.BaseAttack, pokemonProfile.Stats.BaseDefense, pokemonProfile.Stats.BaseStamina)},
			{Name: "**Types**", Value: strings.Join(pokemonProfile.Types.Type, ", ")},
			{Name: "**Moves**", Value: fmt.Sprintf("Charged: %s\nFast: %s", strings.Join(pokemonProfile.Moves.ChargedMoves, ", "), strings.Join(pokemonProfile.Moves.FastMoves, ", "))},
		},
	}

	s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: "Pokémon lookup successful.",
		Embeds:  []*discordgo.MessageEmbed{embed},
	})
}
