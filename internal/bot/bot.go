package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"

	"pogo-bot/internal/api"
	"pogo-bot/internal/commands"
	"pogo-bot/internal/config"
	"pogo-bot/internal/scheduler"
)

// Bot is the long-lived Discord session plus the two engines from the README:
//   1. Interactive Event Pipeline  — slash-command routing
//   2. Autonomous Background Routine — daily 08:00 announcement ticker
type Bot struct {
	Session   *discordgo.Session
	Config    *config.Config
	Commands  *commands.Handler
	Scheduler *scheduler.Scheduler
}

func New(cfg *config.Config) (*Bot, error) {

	session, err := discordgo.New("Bot " + cfg.DiscordToken)

	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}

	session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuilds

	client := api.New()
	cmds := commands.New(client)

	b := &Bot{
		Config:   cfg,
		Commands: cmds,
		Session: session,
		Scheduler: scheduler.New(session, cfg.AnnouncementChannelID, client),
	}

	return b, nil
}

