package main

import (
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

func onGuildMemberJoin(event *events.GuildMemberJoin) {
	if event.GuildID.String() != staticConfig.GuildID {
		return
	}
	go updateVoiceLimit(event.Client())
	var roles []string
	if event.Member.User.Bot {
		roles = staticConfig.JoinRoles.Bots
	} else {
		roles = staticConfig.JoinRoles.Users
	}
	for _, roleId := range roles {
		go func() {
			if err := event.Client().Rest().AddMemberRole(event.GuildID, event.Member.User.ID, snowflake.MustParse(roleId)); err != nil {
				slog.Error("error adding role",
					slog.Any("error", err),
					slog.Any("guild_id", event.GuildID),
					slog.Any("user_id", event.Member.User.ID),
					slog.Any("role_id", roleId),
				)
			}
		}()
	}
	if _, err := event.Client().Rest().CreateMessage(snowflake.MustParse(staticConfig.UserLogChannel), discord.MessageCreate{
		Content: fmt.Sprintf("<:Join:1264245344129253456> %s joined the server.", event.Member.Mention()),
		AllowedMentions: &discord.AllowedMentions{
			Parse: []discord.AllowedMentionType{},
		},
	}); err != nil {
		slog.Error("error logging user join", slog.Any("error", err),
			slog.Any("guild_id", event.GuildID),
			slog.Any("user_id", event.Member.User.ID),
		)
	}

}

func onGuildMemberLeave(event *events.GuildMemberLeave) {
	if event.GuildID.String() != staticConfig.GuildID {
		return
	}
	go updateVoiceLimit(event.Client())
	if _, err := event.Client().Rest().CreateMessage(snowflake.MustParse(staticConfig.UserLogChannel), discord.MessageCreate{
		Content: fmt.Sprintf("<:Leave:1264245459493322823> %s left the server.", event.Member.Mention()),
		AllowedMentions: &discord.AllowedMentions{
			Parse: []discord.AllowedMentionType{},
		},
	}); err != nil {
		slog.Error("error logging user join", slog.Any("error", err),
			slog.Any("guild_id", event.GuildID),
			slog.Any("user_id", event.Member.User.ID),
		)
	}
}

func updateVoiceLimit(client bot.Client) {
	guild, found := client.Caches().Guild(snowflake.MustParse(staticConfig.GuildID))
	if !found {
		return
	}
	client.Rest().UpdateChannel(snowflake.MustParse(staticConfig.VoiceLimitChannel), discord.GuildVoiceChannelUpdate{
		UserLimit: &guild.MemberCount,
	})
}
