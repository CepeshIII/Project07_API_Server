package mailer

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendGridMailer implements the Client interface for sending emails via SendGrid.
type SendGridMailer struct {
	fromEmail string
	apiKey    string
	client    *sendgrid.Client
	isSandbox bool
}

// NewSendgrid creates and returns a new SendGridMailer instance.
func NewSendgrid(apiKey, fromEmail string, isSandbox bool) *SendGridMailer {
	client := sendgrid.NewSendClient(apiKey)

	return &SendGridMailer{
		fromEmail: fromEmail,
		apiKey:    apiKey,
		client:    client,
		isSandbox: isSandbox,
	}
}

// Send delivers an email message via the SendGrid API client.
func (m *SendGridMailer) Send(ctx context.Context, username, userEmail string, compiledTemplate CompiledTemplate) error {
	from := mail.NewEmail(FromName, m.fromEmail)
	to := mail.NewEmail(username, userEmail)
	message := mail.NewSingleEmail(from, compiledTemplate.Subject, to, "", compiledTemplate.Body)

	message.SetMailSettings(
		&mail.MailSettings{
			SandboxMode: &mail.Setting{
				Enable: &m.isSandbox,
			},
		},
	)

	response, err := m.client.SendWithContext(ctx, message)

	// Network / SDK error
	if err != nil {
		return errors.Join(ErrEmailDeliveryFailed, err)
	}

	// API error status check (catch all client and server errors)
	if response.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("%w: status %d (%s)", ErrEmailDeliveryFailed, response.StatusCode, response.Body)
	}

	return nil
}
