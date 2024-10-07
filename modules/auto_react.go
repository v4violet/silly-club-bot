//go:build modules.all || modules.auto_react

package modules

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"go.uber.org/fx"
)

//go:embed auto_react.json
var auto_reactions_raw string

var auto_reactions = map[*regexp.Regexp]string{}

func init() {
	modules = append(modules, fx.Module("modules/auto_react",
		fx.Provide(ProvideAutoReact),
		fx.Invoke(NewAutoReact),
	))
}

func ProvideAutoReact() Results {
	return Results{
		Options: []bot.ConfigOpt{
			bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuildMessages, gateway.IntentMessageContent)),
		},
	}
}

func ProcessAutoReactions() error {
	auto_reactions_unprocessed := map[string]string{}
	if err := json.Unmarshal([]byte(auto_reactions_raw), &auto_reactions_unprocessed); err != nil {
		return err
	}
	for regex_raw, emojiId := range auto_reactions_unprocessed {
		regex, err := regexp.Compile(regex_raw)
		if err != nil {
			slog.Error("failed to compile regex",
				slog.Any("error", err),
				slog.String("regex", regex_raw),
				slog.String("emoji_id", emojiId),
			)
			return err
		}
		auto_reactions[regex] = emojiId
	}
	return nil
}

func NewAutoReact(p Params) error {
	if err := ProcessAutoReactions(); err != nil {
		return err
	}
	p.Client.AddEventListeners(
		bot.NewListenerFunc(func(event *events.GuildMessageCreate) {
			if event.Message.Author.Bot || event.GuildID != p.DiscordConfig.GuildId {
				return
			}
			parse_time := time.Duration(0)
			matched := 0
			for regex, emojiId := range auto_reactions {
				start := time.Now()
				if !regex.MatchString(event.Message.Content) {
					continue
				}
				parse_time += time.Since(start)
				matched += 1
				go func() {
					if err := event.Client().Rest().AddReaction(event.ChannelID, event.MessageID, emojiId); err != nil {
						slog.Error("error adding reaction",
							slog.Any("error", err),
							slog.Any("channel_id", event.ChannelID),
							slog.Any("message_id", event.MessageID),
						)
					}
				}()
			}
			if matched > 0 {
				slog.Debug("matched triggers", slog.Int("matches", matched), slog.Any("duration", parse_time))
				if strings.Index(strings.ToLower(event.Message.Content), "-debug") != -1 {
					if _, err := event.Client().Rest().CreateMessage(event.ChannelID, discord.MessageCreate{
						Content: fmt.Sprintf("matches: %v\nduration: %v", matched, parse_time),
						MessageReference: &discord.MessageReference{
							MessageID: &event.MessageID,
							ChannelID: &event.ChannelID,
							GuildID:   &event.GuildID,
						},
					}); err != nil {
						slog.Error("error sending debug message",
							slog.Any("error", err),
							slog.Any("channel_id", event.ChannelID),
							slog.Any("message_id", event.MessageID),
						)
					}
				}
			}
		}),
	)
	return nil
}
