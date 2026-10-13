// Package mailer provides email templates, embedding assets, and client interfaces
// for sending application notifications and user invitations.
package mailer

import (
	"context"
	"embed"
	"errors"
)

const (
	// FromName is the display name used in the sender field of outgoing emails.
	FromName = "Project07 API Server"
	// MaxRetries defines the maximum number of delivery attempts for a failed email.
	MaxRetries = 1
	// UserWelcomeTemplate specifies the filename for the user invitation email template.
	UserWelcomeTemplate = "user_invitation.tmpl"
)

// ErrEmailDeliveryFailed is returned when an email fails to deliver after all retry attempts.
var ErrEmailDeliveryFailed = errors.New("failed to deliver email")

// FS embeds the email templates directory into the binary.
//
//go:embed "templates"
var FS embed.FS

// Client defines the interface for email delivery providers (e.g., SendGrid, Mailtrap).
type Client interface {
	Send(ctx context.Context, username, userEmail string, compiledTemplate CompiledTemplate) error
}
