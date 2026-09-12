package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func formatMoves(moves []string) string {
	if len(moves) == 0 {
		return "None"
	}
	return strings.Join(moves, ", ")
}

// Lookup handles /pokemon-lookup.
func (h *Handler) Lookup(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var targetPokemon, targetForm string
	for _, option := range i.ApplicationCommandData().Options {
		switch option.Name {
		case "pokemon-name":
			targetPokemon = option.StringValue()
		case "form":
			targetForm = option.StringValue()
		}
	}
	if targetPokemon == "" {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Please provide a valid Pokémon name.",
		})
		return
	}
	pokemonProfile, err := h.API.LookupPokemon(targetPokemon, targetForm)
	if err != nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: err.Error(),
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
			{Name: "**Moves**", Value: fmt.Sprintf("Fast: %s\nCharged: %s\nElite Fast: %s\nElite Charged: %s", formatMoves(pokemonProfile.Moves.FastMoves), formatMoves(pokemonProfile.Moves.ChargedMoves), formatMoves(pokemonProfile.Moves.EliteFastMoves), formatMoves(pokemonProfile.Moves.EliteChargedMoves))},
		},
	}
	if pokemonProfile.GOImage != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: pokemonProfile.GOImage}
	}

	s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: "Pokémon lookup successful.",
		Embeds:  []*discordgo.MessageEmbed{embed},
	})
}
