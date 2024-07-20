package main

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func onGuildVoiceStateUpdate(event *events.GuildVoiceStateUpdate) {
	if event.VoiceState.GuildID.String() != staticConfig.GuildID {
		return
	}
	if event.VoiceState.ChannelID != nil {
		if event.VoiceState.ChannelID != event.OldVoiceState.ChannelID {
			if event.OldVoiceState.ChannelID != nil {
				event.Client().Rest().CreateMessage(*event.OldVoiceState.ChannelID, discord.MessageCreate{
					Content: fmt.Sprintf("<:Leave:1236848876879741060> %s moved to <#%s>.", event.Member.Mention(), *event.VoiceState.ChannelID),
					AllowedMentions: &discord.AllowedMentions{
						Parse: []discord.AllowedMentionType{},
					},
				})
			}
			event.Client().Rest().CreateMessage(*event.OldVoiceState.ChannelID, discord.MessageCreate{
				Content: fmt.Sprintf("<:Leave:1236848876879741060> %s moved to <#%s>.", event.Member.Mention(), *event.VoiceState.ChannelID),
				AllowedMentions: &discord.AllowedMentions{
					Parse: []discord.AllowedMentionType{},
				},
			})
		} else if event.VoiceState.SessionID != event.OldVoiceState.SessionID {
			joinEnd := "the channel"
			if event.OldVoiceState.ChannelID != nil {
				joinEnd = fmt.Sprintf("from <#%s>", *event.OldVoiceState.ChannelID)
			}
			event.Client().Rest().CreateMessage(*event.OldVoiceState.ChannelID, discord.MessageCreate{
				Content: fmt.Sprintf("<:Join:1236848875919249429> %s joined %s.", event.Member.Mention(), joinEnd),
				AllowedMentions: &discord.AllowedMentions{
					Parse: []discord.AllowedMentionType{},
				},
			})
		}
	} else if event.VoiceState.ChannelID != nil && event.OldVoiceState.ChannelID != nil {
		event.Client().Rest().CreateMessage(*event.OldVoiceState.ChannelID, discord.MessageCreate{
			Content: fmt.Sprintf("<:Leave:1264245459493322823> %s left the channel.", event.Member.Mention()),
			AllowedMentions: &discord.AllowedMentions{
				Parse: []discord.AllowedMentionType{},
			},
		})
	}
}
