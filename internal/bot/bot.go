package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"pogo-bot/internal/api"
	"pogo-bot/internal/cache"
	"pogo-bot/internal/commands"
	"pogo-bot/internal/config"
	"pogo-bot/internal/scheduler"
)

// Bot is the long-lived Discord session plus the two engines from the README:
//  1. Interactive Event Pipeline  — slash-command routing
//  2. Autonomous Background Routine — daily 08:00 announcement ticker
type Bot struct {
	Session   *discordgo.Session
	Config    *config.Config
	Commands  *commands.Handler
	Scheduler *scheduler.Scheduler
	Cache     *cache.Cache
}

func New(cfg *config.Config) (*Bot, error) {

	session, err := discordgo.New("Bot " + cfg.DiscordToken)

	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}

	session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuilds

	rdb, err := cache.New(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis cache: %w", err)
	}
	client := api.New(rdb)
	cmds := commands.New(client, rdb)

	b := &Bot{
		Config:    cfg,
		Commands:  cmds,
		Session:   session,
		Scheduler: scheduler.New(session, cfg.AnnouncementChannelID, client, rdb),
		Cache:     rdb,
	}

	b.registerInteractionCreateHandler(session, cmds)

	return b, nil
}

func (b *Bot) registerInteractionCreateHandler(session *discordgo.Session, commands *commands.Handler) {
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		commands.Route(s, i)
	})
}

func (b *Bot) Start() error {
	if err := b.Session.Open(); err != nil {
		return fmt.Errorf("failed to open Discord session: %w", err)
	}
	_, err := b.Session.ApplicationCommandBulkOverwrite(
		b.Session.State.User.ID,
		"",
		b.Commands.Definitions(),
	)
	if err != nil {
		return fmt.Errorf("failed to register slash commands: %w", err)
	}
	go b.Scheduler.Run()
	return nil
}

func (b *Bot) Close() {
	b.Scheduler.Stop()
	b.Cache.Close()
	b.Session.Close()
}
