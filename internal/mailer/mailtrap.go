package mailer

import (
	"context"
	"errors"
	"fmt"

	"github.com/mailtrap/mailtrap-go"
)

const (
	// mailtrapSandboxID defines the default sandbox ID configuration for Mailtrap testing.
	mailtrapSandboxID = 4840248
)

var errMailtrapRequestFailed = errors.New("mailtrap provider error: request was not successful")

// MailtrapMailer implements the Client interface for sending emails via Mailtrap.
type MailtrapMailer struct {
	fromEmail string
	apiKey    string
	client    *mailtrap.Client
}

// NewMailtrap creates and returns a new MailtrapMailer instance.
func NewMailtrap(apiKey, fromEmail string, isSandbox bool) (*MailtrapMailer, error) {
	client, err := mailtrap.NewClient(apiKey,
		mailtrap.WithSandbox(isSandbox),
		mailtrap.WithSandboxID(mailtrapSandboxID),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create mailtrap client: %w", err)
	}

	return &MailtrapMailer{
		fromEmail: fromEmail,
		apiKey:    apiKey,
		client:    client,
	}, nil
}

// Send delivers an email message via the Mailtrap API client.
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
		return errors.Join(errMailtrapRequestFailed, err)
	}

	if !response.Success {
		return errMailtrapRequestFailed
	}

	return nil
}
