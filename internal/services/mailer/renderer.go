package mailer

import (
	"bytes"
	"html/template"
)

type Renderer struct {
	tmpl *template.Template
}

func NewRenderer(pattern string) (*Renderer, error) {
	t, err := template.ParseGlob(pattern)
	if err != nil {
		return nil, err
	}

	return &Renderer{tmpl: t}, nil
}

func (r *Renderer) Render(name string, data any) (string, error) {
	var buf bytes.Buffer

	err := r.tmpl.ExecuteTemplate(&buf, name, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
