package voice_log

import (
	"log/slog"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/v4violet/silly-club-bot/config"
	"github.com/v4violet/silly-club-bot/modules"
	"github.com/v4violet/silly-club-bot/templates"
)

func Init() {
	modules.RegisterModule(modules.Module{
		Name: "voice_log",
		Init: func() []bot.ConfigOpt {
			return []bot.ConfigOpt{
				bot.WithEventListenerFunc(onGuildVoiceStateUpdate),
				bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuildVoiceStates)),
				bot.WithCacheConfigOpts(
					cache.WithCaches(cache.FlagVoiceStates),
				),
			}
		},
	})
}

type VoiceStateTemplateData struct {
	User            string
	Channel         *string
	PreviousChannel *string
}

func onGuildVoiceStateUpdate(event *events.GuildVoiceStateUpdate) {
	if event.VoiceState.GuildID != config.Config.Discord.GuildId {
		return
	}
	templateData := VoiceStateTemplateData{
		User:            event.Member.Mention(),
		Channel:         nil,
		PreviousChannel: nil,
	}
	if event.VoiceState.ChannelID != nil {
		mention := discord.ChannelMention(*event.VoiceState.ChannelID)
		templateData.Channel = &mention
	}
	if event.OldVoiceState.ChannelID != nil {
		mention := discord.ChannelMention(*event.OldVoiceState.ChannelID)
		templateData.PreviousChannel = &mention
	}

	isInVoice := event.VoiceState.ChannelID != nil
	wasInVoice := event.OldVoiceState.ChannelID != nil
	sameChannel := false
	sameSession := event.OldVoiceState.SessionID == event.VoiceState.SessionID

	if isInVoice && wasInVoice {
		sameChannel = (*event.OldVoiceState.ChannelID) == (*event.VoiceState.ChannelID)
	}

	// (!wasInVoice or !sameChannel) and isInVoice = join
	// wasInVoice and !isInVoice = leave
	// wasInVoice and isInVoice and !sameChannel = move
	// sameChannel and !sameSession = rejoin
	if (!wasInVoice || !sameChannel) && isInVoice {
		if _, err := event.Client().Rest().CreateMessage(*event.VoiceState.ChannelID, discord.MessageCreate{
			Content: templates.Exec("modules.voice_log.join", templateData),
			AllowedMentions: &discord.AllowedMentions{
				Parse: []discord.AllowedMentionType{},
			},
		}); err != nil {
			slog.Error("error sending voice join message",
				slog.Any("error", err),
				slog.Any("channel_id", *event.VoiceState.ChannelID),
			)
		}
	}
	if wasInVoice && !isInVoice {
		// leave
		if _, err := event.Client().Rest().CreateMessage(*event.OldVoiceState.ChannelID, discord.MessageCreate{
			Content: templates.Exec("modules.voice_log.leave", templateData),
			AllowedMentions: &discord.AllowedMentions{
				Parse: []discord.AllowedMentionType{},
			},
		}); err != nil {
			slog.Error("error sending voice leave message",
				slog.Any("error", err),
				slog.Any("channel_id", *event.OldVoiceState.ChannelID),
			)
		}
	}
	if wasInVoice && isInVoice {
		if !sameChannel {
			// move
			if _, err := event.Client().Rest().CreateMessage(*event.OldVoiceState.ChannelID, discord.MessageCreate{
				Content: templates.Exec("modules.voice_log.move", templateData),
				AllowedMentions: &discord.AllowedMentions{
					Parse: []discord.AllowedMentionType{},
				},
			}); err != nil {
				slog.Error("error sending voice move message",
					slog.Any("error", err),
					slog.Any("channel_id", *event.OldVoiceState.ChannelID),
				)
			}
		} else if !sameSession {
			// rejoin
			if _, err := event.Client().Rest().CreateMessage(*event.VoiceState.ChannelID, discord.MessageCreate{
				Content: templates.Exec("modules.voice_log.rejoin", templateData),
				AllowedMentions: &discord.AllowedMentions{
					Parse: []discord.AllowedMentionType{},
				},
			}); err != nil {
				slog.Error("error sending voice rejoin message",
					slog.Any("error", err),
					slog.Any("channel_id", *event.VoiceState.ChannelID),
				)
			}
		}
	}
}
