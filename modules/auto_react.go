//go:build modules.all || modules.auto_react

package modules

import (
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

var auto_reactions_raw = map[string]string{
	"(?i)(penis|cock|dick|cum)":      "cum:1144596862075154562",
	"(?i)(ch|chapter\\s*)2":          "ch2_wen:1144404092093997056",
	"(?i)(boy(s|kisser)|yoai)":       "boykisser:1156664341286899772",
	"(?i)(girl(s|kisser)|yuri)":      "girlkisser:1202306410352738354",
	"(?i)cop(e|i(um|ng))":            "COPIUM:1144404181000671354",
	"(?i)cucumber":                   "cucumber:1237250194089971712",
	"(?i)pipe":                       "metalPipe:1236853099360948255",
	"(?i)(bean|🫘)":                   "🫘",
	"(?i)(hl|half(-|\\s)+life\\s*)3": "hl3_wen:1271947068503756872",
	"(?i)ivy":                        "ivykisser:1280588649276244038",
	"(?i)trans":                      "transgender:1195201358383554681",
	"(?i)horny":                      "panting:1144606367924097124",
}

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

func NewAutoReact(p Params) error {
	for regex_raw, emojiId := range auto_reactions_raw {
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
