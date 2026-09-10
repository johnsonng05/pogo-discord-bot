package commands

import (
	"context"
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
)

// SetAnnounceChannel handles /pogo-set-announce-channel.
func (h *Handler) SetAnnounceChannel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if h.Cache == nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Redis Cache not configured.",
		})
		return
	}

	if i.GuildID == "" {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "This command must be used in a server.",
		})
		return
	}

	if i.Member == nil || i.Member.Permissions&discordgo.PermissionManageServer == 0 {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "You do not have permission to use this command.",
		})
		return
	}

	existingChannelID, err := h.Cache.GetAnnouncementChannel(context.Background(), i.GuildID)
	if err != nil && !errors.Is(err, redis.Nil) {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Failed to read announcement channel.",
		})
		return
	}

	if existingChannelID == i.ChannelID {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("Daily announcements are already posted in <#%s>.", existingChannelID),
		})
		return
	}

	if err := h.Cache.SetAnnouncementChannel(context.Background(), i.GuildID, i.ChannelID); err != nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Failed to set announcement channel.",
		})
		return
	}

	s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: fmt.Sprintf("Daily announcements will post in <#%s> at 10:00 AM PT.", i.ChannelID),
	})
}

func (h *Handler) StopAnnounceChannel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if h.Cache == nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Redis Cache not configured.",
		})
		return
	}

	if i.GuildID == "" {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "This command must be used in a server.",
		})
		return
	}

	if i.Member == nil || i.Member.Permissions&discordgo.PermissionManageServer == 0 {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "You do not have permission to use this command.",
		})
		return
	}

	existingChannelID, err := h.Cache.GetAnnouncementChannel(context.Background(), i.GuildID)
	if err != nil && !errors.Is(err, redis.Nil) {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Failed to read announcement channel.",
		})
		return
	}

	if errors.Is(err, redis.Nil) || existingChannelID == "" {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Daily announcements are not configured for this server.",
		})
		return
	}

	if err := h.Cache.DeleteAnnouncementChannel(context.Background(), i.GuildID); err != nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Failed to delete announcement channel.",
		})
		return
	}

	s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: fmt.Sprintf("Daily announcements will no longer post in <#%s>.", existingChannelID),
	})
}
