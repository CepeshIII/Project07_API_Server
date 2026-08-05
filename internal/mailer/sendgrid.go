package mailer

import (
	"bytes"
	"fmt"
	"log"
	"text/template"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridMailer struct {
	fromEmail string
	apiKey    string
	client    *sendgrid.Client
}

func NewSendgrid(apiKey, fromEmail string) *SendGridMailer {
	client := sendgrid.NewSendClient(apiKey)

	return &SendGridMailer{
		fromEmail: fromEmail,
		apiKey:    apiKey,
		client:    client,
	}
}

func (m *SendGridMailer) Send(templateFile, username, email string, data any, isSandbox bool) error {
	from := mail.NewEmail(FromName, m.fromEmail)
	to := mail.NewEmail(username, email)

	// template parsing and building
	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return fmt.Errorf("failed to execute subject template: %w", err)
	}

	body := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(body, "body", data)
	if err != nil {
		return fmt.Errorf("failed to execute body template: %w", err)
	}

	message := mail.NewSingleEmail(from, subject.String(), to, "", body.String())

	message.SetMailSettings(
		&mail.MailSettings{
			SandboxMode: &mail.Setting{
				Enable: &isSandbox,
			},
		},
	)

	for i := range MaxRetries {
		response, err := m.client.Send(message)
		if err != nil {
			log.Printf("failed to send email to %v, attempt: %d of %d", email, i+1, MaxRetries)
			log.Printf("Error: %v", err.Error())

			// exponential backoff before retrying
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		if response.StatusCode >= 300 {
			log.Printf("Bad response to send email to %v, attempt: %d of %d", email, i+1, MaxRetries)
			log.Printf("Response code: %v", response.StatusCode)
			log.Printf("Response body: %s", response.Body)

			// exponential backoff before retrying
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		log.Printf("Email sent to %v, status code: %d", email, response.StatusCode)

		return nil
	}

	return fmt.Errorf("failed to send email to %v after %d attempts", email, MaxRetries)
}
