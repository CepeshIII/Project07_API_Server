package main

import (
	"context"
	"fmt"
	"time"

	"github.com/CepeshIII/Project07_API_Server/internal/credential"
	"github.com/CepeshIII/Project07_API_Server/internal/mailer"
	"github.com/CepeshIII/Project07_API_Server/internal/store"
)

func (app *application) sendUserActivation(ctx context.Context, plainToken string, userWithRole *store.UserWithRole) error {
	activationURL := fmt.Sprintf("%s/confirm/%s", app.config.FrontendURL, plainToken)

	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      userWithRole.Username,
		ActivationURL: activationURL,
	}

	compiledTemplate, err := mailer.BuildTemplate(mailer.UserWelcomeTemplate, vars)
	if err != nil {
		return err
	}

	var (
		lastSendErr error
		successSend bool
	)

	for i := range app.config.Mail.MaxRetries {
		lastSendErr = app.mailer.Send(ctx, userWithRole.Username, userWithRole.Email, compiledTemplate)
		if lastSendErr == nil {
			successSend = true

			break
		}

		app.logger.Warnw("failed to send activation email",
			"attempt", i+1,
			"max_retries", app.config.Mail.MaxRetries,
			"email", userWithRole.Email,
			"error", lastSendErr,
		)

		// exponential backoff before retrying
		time.Sleep(time.Second * time.Duration(i+1))

		continue
	}

	// send the invitation email
	if !successSend {
		// rollback the user creation if email sending fails (SAGA pattern)
		rollbackErr := app.store.Users.DeleteUserAndInvitation(ctx, userWithRole.ID)
		if rollbackErr != nil {
			app.logger.Errorw("error rolling back user creation", "error", rollbackErr)
		}

		return lastSendErr
	}

	return nil
}

func (app *application) Register(ctx context.Context, payload RegisterUserPayload) (*store.User, string, error) {
	userWithRole := &store.UserWithRole{
		User: store.User{
			Username: payload.Username,
			Email:    payload.Email,
		},
		Role: store.Role{
			Name: "user",
		},
	}

	if err := userWithRole.Password.Set(payload.Password); err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	plainToken, tokenHash := credential.New()

	if err := app.store.Users.CreateAndInvite(ctx, userWithRole, tokenHash, app.config.Mail.Exp); err != nil {
		return nil, "", err // Caught by handler for duplicate checks
	}

	if err := app.sendUserActivation(ctx, plainToken, userWithRole); err != nil {
		return nil, "", fmt.Errorf("failed to send activation email: %w", err)
	}

	return &userWithRole.User, plainToken, nil
}
