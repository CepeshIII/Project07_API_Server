package mailer

import (
	"context"
	"fmt"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridMailer struct {
	fromEmail string
	apiKey    string
	client    *sendgrid.Client
	isSandbox bool
}

func NewSendgrid(apiKey, fromEmail string, isSandbox bool) *SendGridMailer {
	client := sendgrid.NewSendClient(apiKey)

	return &SendGridMailer{
		fromEmail: fromEmail,
		apiKey:    apiKey,
		client:    client,
		isSandbox: isSandbox,
	}
}

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
		return fmt.Errorf("%w: %v", ErrEmailDeliveryFailed, err)
	}

	// API status >= 300
	if response.StatusCode == 300 {
		return fmt.Errorf("%w: status %d (%s)", ErrEmailDeliveryFailed, response.StatusCode, response.Body)
	}

	return nil
}
