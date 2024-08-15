package templates

import (
	"embed"
	"errors"
	"fmt"
	"text/template"

	"github.com/v4violet/silly-club-bot/emojis"
	"go.uber.org/fx"
)

//go:embed *
var templatesFs embed.FS

var Module = fx.Module("templates", fx.Provide(NewTemplates))

func NewTemplates(emojis *emojis.Emojis) (*template.Template, error) {
	tmpl, err := template.ParseFS(templatesFs, "*/*.tmpl")
	if err != nil {
		return nil, errors.Join(errors.New("failed to parsefs templates"), err)
	}

	for k, v := range *emojis {
		emoji_tmpl, err := tmpl.New(fmt.Sprintf("emojis.%s", k)).Parse(v.Discord.Mention())
		if err != nil {
			return nil, err
		}
		tmpl = emoji_tmpl
	}

	return tmpl, nil
}
