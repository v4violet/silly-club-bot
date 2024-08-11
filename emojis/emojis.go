package emojis

import (
	"embed"
	"io/fs"
	"path"
	"strings"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
)

//go:embed *
var emojiFs embed.FS

type Emoji struct {
	Discord *discord.Emoji
	Entry   fs.DirEntry
}

var Emojis = map[string]*Emoji{}

func Load(client bot.Client) error {
	emojis_dir, err := emojiFs.ReadDir("./")
	if err != nil {
		return err
	}

	for _, entry := range emojis_dir {
		if entry.Type().IsRegular() {
			Emojis[strings.TrimSuffix(entry.Name(), path.Ext(entry.Name()))] = &Emoji{
				Entry: entry,
			}
		}
	}

	application_emojis, err := client.Rest().GetApplicationEmojis(client.ApplicationID())

	if err != nil {
		return err
	}

	for _, application_emoji := range application_emojis {
		if emoji, ok := Emojis[application_emoji.Name]; ok {
			emoji.Discord = &application_emoji
		}
	}

	for k, v := range Emojis {
		if v.Discord != nil {
			continue
		}
		file, err := emojiFs.Open(v.Entry.Name())
		if err != nil {
			return err
		}

		icon, err := discord.NewIcon(discord.IconTypeUnknown, file)
		if err != nil {
			return err
		}

		emoji, err := client.Rest().CreateApplicationEmoji(client.ApplicationID(), discord.EmojiCreate{
			Name:  k,
			Image: *icon,
		})
		if err != nil {
			return err
		}
		v.Discord = emoji
	}

	return nil
}
