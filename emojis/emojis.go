package emojis

import (
	"embed"
	"errors"
	"io/fs"
	"path"
	"strings"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"go.uber.org/fx"
)

//go:embed *
var emojiFs embed.FS

type Emoji struct {
	Discord *discord.Emoji
	Entry   fs.DirEntry
}

type Emojis map[string]*Emoji

var Module = fx.Module("emojis", fx.Provide(NewEmojis))

var cachedEmojis *Emojis

func NewEmojis(client bot.Client) (*Emojis, error) {
	emojis_dir, err := emojiFs.ReadDir(".")
	if err != nil {
		return nil, errors.Join(errors.New("error opening embeded directory"), err)
	}

	if cachedEmojis != nil {
		return cachedEmojis, nil
	}

	emojis := make(Emojis, len(emojis_dir))

	for _, entry := range emojis_dir {
		if !entry.Type().IsRegular() {
			continue
		}
		ext := path.Ext(entry.Name())
		switch ext {
		case ".png", ".jpg", ".jpeg", ".gif":
			emojis[strings.TrimSuffix(entry.Name(), ext)] = &Emoji{
				Entry: entry,
			}
		}
	}

	application_emojis, err := client.Rest().GetApplicationEmojis(client.ApplicationID())

	if err != nil {
		return nil, errors.Join(errors.New("error getting application emojis"), err)
	}

	for _, application_emoji := range application_emojis {
		if emoji, ok := emojis[application_emoji.Name]; ok {
			emoji.Discord = &application_emoji
		}
	}

	for k, v := range emojis {
		if v.Discord != nil {
			continue
		}
		file, err := emojiFs.Open(v.Entry.Name())
		if err != nil {
			return nil, errors.Join(errors.New("error opening embeded file"), err)
		}

		var icon_type discord.IconType

		switch path.Ext(v.Entry.Name()) {
		case ".png":
			icon_type = discord.IconTypePNG
		case ".jpg", ".jpeg":
			icon_type = discord.IconTypeJPEG
		case ".gif":
			icon_type = discord.IconTypeGIF
		default:
			icon_type = discord.IconTypeUnknown
		}

		icon, err := discord.NewIcon(icon_type, file)
		if err != nil {
			return nil, errors.Join(errors.New("error creating new discord icon"), err)
		}

		emoji, err := client.Rest().CreateApplicationEmoji(client.ApplicationID(), discord.EmojiCreate{
			Name:  k,
			Image: *icon,
		})
		if err != nil {
			return nil, errors.Join(errors.New("error creating application emoji"), err)
		}
		v.Discord = emoji
	}

	cachedEmojis = &emojis
	return cachedEmojis, nil
}
