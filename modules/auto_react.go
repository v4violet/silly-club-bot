//go:build modules.all || modules.auto_react

package modules

import (
	"log/slog"
	"regexp"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/v4violet/silly-club-bot/config"
)

var auto_reactions_raw = map[string]string{
	"(?i)(penis|cock|dick|cum)": "cum:1144596862075154562",
	"(?i)(ch|chapter )2":        "ch2_wen:1144404092093997056",
	"(?i)boy(s|kisser)":         "boykisser:1156664341286899772",
	"(?i)girl(s|kisser)":        "girlkisser:1202306410352738354",
	"(?i)cope":                  "COPIUM:1144404181000671354",
	"(?i)cucumber":              "cucumber:1237250194089971712",
	"(?i)sus":                   "amogus:1144404244615659590",
	"(?i)pipe":                  "metalPipe:1236853099360948255",
	"(?i)bean":                  "🫘",
}

var auto_reactions = map[*regexp.Regexp]string{}

func init() {
	Modules["auto_react"] = Module{
		Init: func() ([]bot.ConfigOpt, error) {
			for regex_raw, emojiId := range auto_reactions_raw {
				regex, err := regexp.Compile(regex_raw)
				if err != nil {
					slog.Error("failed to compile regex",
						slog.Any("error", err),
						slog.String("regex", regex_raw),
						slog.String("emoji_id", emojiId),
					)
					return nil, err
				}
				auto_reactions[regex] = emojiId
			}

			return []bot.ConfigOpt{
				bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuildMessages, gateway.IntentMessageContent)),
				bot.WithEventListenerFunc(func(event *events.GuildMessageCreate) {
					if event.Message.Author.Bot || event.GuildID != config.Config.Discord.GuildId {
						return
					}
					for regex, emojiId := range auto_reactions {
						if !regex.MatchString(event.Message.Content) {
							continue
						}
						if err := event.Client().Rest().AddReaction(event.ChannelID, event.MessageID, emojiId); err != nil {
							slog.Error("error adding reaction",
								slog.Any("error", err),
								slog.Any("channel_id", event.ChannelID),
								slog.Any("message_id", event.MessageID),
							)
						}
					}
				}),
			}, nil
		},
	}
}
