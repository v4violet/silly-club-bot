package main

import (
	"log/slog"
	"math/rand"
	"regexp"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

func onMessageCreate(event *events.MessageCreate) {
	if event.GuildID.String() != staticConfig.GuildID {
		return
	}
	var rest = event.Client().Rest()
	for regex, emojiId := range staticConfig.AutoReact {
		go func(regex *regexp.Regexp, emojiId string) {
			if regex.MatchString(event.Message.Content) {
				if err := rest.AddReaction(event.ChannelID, event.MessageID, emojiId); err != nil {
					slog.Error("error adding reaction",
						slog.Any("error", err),
						slog.Any("channel_id", event.ChannelID),
						slog.Any("message_id", event.MessageID),
					)
				}
			}
		}(regex, emojiId)
	}

	_, channelWhitelisted := staticConfig.RandomReaction.ChannelWhitelist[event.ChannelID.String()]

	if rand.Float32() <= (staticConfig.RandomReaction.Percentage/100) && channelWhitelisted {
		go func() {
			if err := rest.AddReaction(event.ChannelID, event.MessageID, staticConfig.RandomReaction.Emoji); err != nil {
				slog.Error("error adding reaction",
					slog.Any("error", err),
					slog.Any("channel_id", event.ChannelID),
					slog.Any("message_id", event.MessageID),
				)
				return
			}
			if _, err := rest.CreateMessage(snowflake.MustParse(staticConfig.RandomReaction.LogChannel), discord.MessageCreate{
				Content: event.Message.JumpURL(),
			}); err != nil {
				slog.Error("error logging random reaction",
					slog.Any("error", err),
					slog.Any("channel_id", event.ChannelID),
					slog.Any("message_id", event.MessageID),
				)
			}
		}()
	}
}

const pinEmoji = "📌"

func onMessageReactionAdd(event *events.MessageReactionAdd) {
	if event.Emoji.Name == nil || event.GuildID.String() != staticConfig.GuildID {
		return
	}
	if *event.Emoji.Name != pinEmoji {
		return
	}
	message, err := event.Client().Rest().GetMessage(event.ChannelID, event.MessageID)
	if err != nil {
		slog.Error("error getting reaction message", slog.Any("message_id", message.ID))
		return
	}
	if message.Pinned {
		return
	}
	count := 0
	for _, reaction := range message.Reactions {
		if reaction.Emoji.Name == pinEmoji {
			count = reaction.Count
		}
	}
	if count >= staticConfig.VotePinMinVotes {
		event.Client().Rest().PinMessage(event.ChannelID, event.MessageID)
	}
}
