package commands

import (
	"github.com/bwmarrin/discordgo"

	"pogo-bot/internal/api"
	"pogo-bot/internal/cache"
)

// Command names registered with Discord. Keep these stable — changing a name
// after users have the command in their client is a breaking change.
const (
	NameCurrentEvents  = "pogo-current-events"
	NameUpcomingEvents = "pogo-upcoming-events"
	NameRaids          = "pogo-raids"
	NameLookup         = "pokemon-lookup"
)

// Handler owns slash-command definitions and interaction routing.
type Handler struct {
	API   *api.Client
	Cache *cache.Cache
}

func New(client *api.Client, rdb *cache.Cache) *Handler {
	return &Handler{API: client, Cache: rdb}
}

// Definitions returns the global application-command payloads to register
// on startup (see README: /pogo-events, /pogo-raids, /pokemon-lookup).
func (h *Handler) Definitions() []*discordgo.ApplicationCommand {

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        NameCurrentEvents,
			Description: "Get the latest Pokemon Go events",
		},
		{
			Name:        NameUpcomingEvents,
			Description: "Get the upcoming Pokemon Go events",
		},
		{
			Name:        NameRaids,
			Description: "Get the latest Pokemon Go raids",
		},
		{
			Name:        NameLookup,
			Description: "Look up a Pokemon by name",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "pokemon-name",
					Description: "The Pokemon to lookup",
					Required:    true,
				},
			},
		},
	}
	return commands
}

// Route is called from the Discord interaction handler.

func (h *Handler) Route(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Ignore non-application-command interactions
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	// Defer the response to avoid timeouts
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	// If the response fails, return
	if err != nil {
		return
	}

	// Dispatch to the appropriate handler
	switch i.ApplicationCommandData().Name {
	case NameCurrentEvents:
		h.CurrentEvents(s, i)
	case NameUpcomingEvents:
		h.UpcomingEvents(s, i)
	case NameRaids:
		h.Raids(s, i)
	case NameLookup:
		h.Lookup(s, i)
	}
}
