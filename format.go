package linkchecker

import (
	"encoding/json"
	"html/template"
	"sort"
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
	sort.Slice(lc.Links, func(i, j int) bool {
		return lc.Links[i].URL < lc.Links[j].URL
	})
	j, err := json.Marshal(lc.Links)
	if err != nil {
		return "", err
	}
	return string(j), nil
}

func (lt LinksTerminal) Format(lc *LinkChecker) (string, error) {
	sort.Slice(lc.Links, func(i, j int) bool {
		return lc.Links[i].URL < lc.Links[j].URL
	})
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
