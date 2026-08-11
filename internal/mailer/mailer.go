package mailer

import (
	"context"
	"embed"
	"errors"
)

const (
	FromName            = "Project07 API Server"
	MaxRetries          = 1
	UserWelcomeTemplate = "user_invitation.tmpl"
)

var ErrEmailDeliveryFailed = errors.New("failed to deliver email")

//go:embed "templates"
var FS embed.FS

type Client interface {
	Send(ctx context.Context, username, userEmail string, compiledTemplate CompiledTemplate) error
}
