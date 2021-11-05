package linkchecker

import (
	"encoding/json"
	"html/template"
	"strings"
)

type Formatter interface {
	Format(lc *LinkChecker) (string, error)
}

type LinksJSON struct {
}

type LinksTerminal struct {
}

func (lj LinksJSON) Format(lc *LinkChecker) (string, error) {
	j, err := json.Marshal(lc.Links)
	if err != nil {
		return "", err
	}
	return string(j), nil
}

func (lt LinksTerminal) Format(lc *LinkChecker) (string, error) {
	terminal := `{{- range .Links -}}
{{println .Status  .URL}}

{{- end}}`
	tmpl, err := template.New("term").Parse(terminal)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	err = tmpl.Execute(&sb, lc)
	if err != nil {
		return "", err
	}

	return sb.String(), nil

}
