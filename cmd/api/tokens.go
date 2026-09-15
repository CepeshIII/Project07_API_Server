package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/CepeshIII/Project07_API_Server/internal/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var jwtSecret = []byte("your-super-secret-key-from-env") // Секретний ключ сервера

func NewInvitationToken() (string, string, error) {

	plainToken := uuid.New().String()

	return plainToken, HashToken(plainToken), nil
}

func HashToken(token string) string {
	// hash the token for storage but keep the plain token for email
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func NewSessionToken() (string, string, error) {
	return NewInvitationToken()
}

// Claims — структура даних всередині токена
type CustomClaims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(userID int64, role string, exp time.Time) (string, error) {
	claims := CustomClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func GenerateAccessToken1(claims CustomClaims) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// 2. Перевірка Access Token у Middleware (Без БД!)
func ValidateAccessToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Перевіряємо, чи збігається алгоритм підпису
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err // Токен підроблений або прострочений
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil // Токен валідний! Повертаємо дані користувача
	}

	return nil, errors.New("invalid token")
}

func (app *application) generateAccessToken(userID int64, role string) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(app.config.auth.tokens.accessTokenExp)

	claims := auth.CustomClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    app.config.auth.jwtAuth.iss,
			Audience: jwt.ClaimStrings{
				app.config.auth.jwtAuth.iss,
			},
		},
	}

	token, err := app.auth.GenerateToken(claims)
	if err != nil {
		return "", time.Time{}, err
	}

	return token, exp, nil
}
