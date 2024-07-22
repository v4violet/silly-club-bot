package auto_react

import (
	"log/slog"
	"regexp"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/v4violet/silly-club-bot/config"
	"github.com/v4violet/silly-club-bot/modules"
)

var reactionsRaw = map[string]string{
	"(?i)cube(s)?":              "ChallengeCube:1144404212063666237",
	"(?i)ball(s)?":              "ChallengeBall:1144404221815423097",
	"(?i)(penis|cock|dick|cum)": "cum:1144596862075154562",
	"(?i)(ch|chapter )2":        "ch2_wen:1144404092093997056",
	"(?i)boy(s|kisser)":         "boykisser:1156664341286899772",
	"(?i)girl(s|kisser)":        "girlkisser:1202306410352738354",
	"(?i)cope":                  "COPIUM:1144404181000671354",
	"(?i)cucumber":              "cucumber:1237250194089971712",
	"(?i)sus":                   "amogus:1144404244615659590",
	"(?i)pipe":                  "metalPipe:1236853099360948255",
}

var reactions = map[*regexp.Regexp]string{}

func Init() {
	for regexRaw, emojiId := range reactionsRaw {
		regex, err := regexp.Compile(regexRaw)
		if err != nil {
			slog.Warn("failed to compile regex",
				slog.Any("error", err),
				slog.String("regex", regexRaw),
				slog.String("emoji_id", emojiId),
			)
			continue
		}
		reactions[regex] = emojiId
	}

	modules.RegisterModule(modules.Module{
		Name: "auto_react",
		Init: func() []bot.ConfigOpt {
			return []bot.ConfigOpt{
				bot.WithEventListenerFunc(onMessageCreate),
				bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuildMessages, gateway.IntentMessageContent)),
			}
		},
	})
}

func onMessageCreate(event *events.MessageCreate) {
	if event.GuildID == nil || config.Config.Discord.GuildId != *event.GuildID {
		return
	}
	for regex, emojiId := range reactions {
		if regex.MatchString(event.Message.Content) {
			if err := event.Client().Rest().AddReaction(event.ChannelID, event.MessageID, emojiId); err != nil {
				slog.Error("error adding reaction",
					slog.Any("error", err),
					slog.Any("channel_id", event.ChannelID),
					slog.Any("message_id", event.MessageID),
				)
			}
		}
	}
}
