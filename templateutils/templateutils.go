package templateutils

import (
	"bytes"
	"text/template"
)

func MustExecuteTemplateToString(template *template.Template, name string, data any) string {
	var out bytes.Buffer
	err := template.ExecuteTemplate(&out, name, data)
	if err != nil {
		panic(err)
	}
	return out.String()
}
