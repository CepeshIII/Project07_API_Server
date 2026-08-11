package mailer

import (
	"context"
	"fmt"
	"text/template"

	"github.com/mailtrap/mailtrap-go"
)

type MailtrapMailer struct {
	fromEmail        string
	apiKey           string
	client           *mailtrap.Client
	templates        map[string]*template.Template // Pre-compiled templates!
	compiledTemplate CompiledTemplate
}

func NewMailtrap(apiKey, fromEmail string, isSandbox bool) (*MailtrapMailer, error) {
	client, err := mailtrap.NewClient(apiKey,
		mailtrap.WithSandbox(isSandbox),
		mailtrap.WithSandboxID(4840248),
	)

	if err != nil {
		return nil, err
	}

	return &MailtrapMailer{
		fromEmail: fromEmail,
		apiKey:    apiKey,
		client:    client,
	}, nil
}

func (m *MailtrapMailer) Send(ctx context.Context, username, userEmail string, compiledTemplate CompiledTemplate) error {
	request := &mailtrap.SendRequest{
		From:     mailtrap.Address{Email: m.fromEmail, Name: "Mailtrap Test"},
		To:       []mailtrap.Address{{Email: userEmail, Name: username}},
		Subject:  compiledTemplate.Subject,
		HTML:     compiledTemplate.Body,
		Category: "Integration Test",
	}

	// Pass HTTP/Request context down to the mail client
	response, _, err := m.client.Send(ctx, request)
	if err != nil {
		return fmt.Errorf("mailtrap network request failed: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("mailtrap provider error: request was not successful")
	}

	return nil
}
