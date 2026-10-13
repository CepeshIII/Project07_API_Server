package main

import (
	"time"

	"github.com/CepeshIII/Project07_API_Server/internal/auth"
	"github.com/golang-jwt/jwt/v5"
)

func (app *application) generateAccessToken(userID int64, role string) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(app.config.Auth.Tokens.AccessTokenExp)

	claims := auth.CustomClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    app.config.Auth.JWT.Iss,
			Audience: jwt.ClaimStrings{
				app.config.Auth.JWT.Iss,
			},
		},
	}

	token, err := app.auth.GenerateToken(claims)
	if err != nil {
		return "", time.Time{}, err
	}

	return token, exp, nil
}
