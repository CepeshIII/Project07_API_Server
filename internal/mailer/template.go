package mailer

import (
	"bytes"
	"fmt"
	"text/template"
)

type CompiledTemplate struct {
	Subject, Body string
}

func templateParsingAndBuilding(templateFile string, data any) (*bytes.Buffer, *bytes.Buffer, error) {
	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse template: %w", err)
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute subject template: %w", err)
	}

	body := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(body, "body", data)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute body template: %w", err)
	}

	return subject, body, nil
}

func BuildTemplate(templateFile string, data any) (CompiledTemplate, error) {
	subject, body, err := templateParsingAndBuilding(templateFile, data)
	if err != nil {
		return CompiledTemplate{}, fmt.Errorf("building template %s: %w", templateFile, err)
	}

	return CompiledTemplate{
		Subject: subject.String(),
		Body:    body.String(),
	}, nil
}
