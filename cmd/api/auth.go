package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/CepeshIII/Project07_API_Server/internal/mailer"
	"github.com/CepeshIII/Project07_API_Server/internal/store"
)

type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,max=100" default:"user"`
	Email    string `json:"email" validate:"required,email,max=255"  default:"mail@example.com"`
	Password string `json:"password" validate:"required,min=3,max=72"  default:"password"`
}

type RegisterUserResponse struct {
	User  *store.User `json:"user"`
	Token string      `json:"verification_token"`
}

// RegisterUserHandler godoc
//
//	@Summary		Registers a user
//	@Description	Registers a user
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		RegisterUserPayload	true	"User credentials"
//	@Success		201		{object}	RegisterUserResponseEnvelope
//	@Failure		400		{object}	ErrorEnvelope
//	@Failure		409		{object}	ErrorEnvelope
//	@Failure		500		{object}	ErrorEnvelope
//	@Router			/authentication/user [post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
	}

	// hash the user password
	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()

	plainToken, hashToken, err := NewInvitationToken()
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	// store the user
	if err := app.store.Users.CreateAndInvite(ctx, user, hashToken, app.config.mail.exp); err != nil {
		if errors.Is(err, store.ErrorDuplicateEmail) || errors.Is(err, store.ErrorDuplicateUsername) {
			app.conflictResponse(w, r, err)
			return
		}

		app.internalServerError(w, r, err)
		return
	}

	response := &RegisterUserResponse{
		User:  user,
		Token: plainToken,
	}

	activationURL := fmt.Sprintf("%s/confirm/%s", app.config.frontendURL, plainToken)

	isProdEnv := app.config.env == "production"
	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: activationURL,
	}
	// send the invitation email
	err = app.mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, !isProdEnv)
	if err != nil {
		app.logger.Errorw("error sending welcome email", "error", err)

		// rollback the user creation if email sending fails (SAGA pattern)
		if err := app.store.Users.DeleteUserAndInvitation(ctx, user.ID); err != nil {
			app.logger.Errorw("error rolling back user creation", "error", err)
		}

		app.internalServerError(w, r, err)
		return
	}

	// send the response
	if err := app.jsonResponse(w, http.StatusCreated, response); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
