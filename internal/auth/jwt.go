package auth

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrInvalidToken is returned when a token fails structural or cryptographic validation.
	ErrInvalidToken = errors.New("invalid token")

	// ErrUnexpectedSigningMethod is returned when the token's signing method does not match expectations.
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
)

// JWTAuthenticator handles the generation and validation of JSON Web Tokens.
type JWTAuthenticator struct {
	secret []byte
	aud    string
	iss    string
}

// CustomClaims defines the custom payload fields included in the JWT alongside standard registered claims.
type CustomClaims struct {
	jwt.RegisteredClaims

	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
}

// NewJWTAuthenticator initializes a new JWTAuthenticator with the specified secret, audience, and issuer.
func NewJWTAuthenticator(jwtSecret, aud, iss string) *JWTAuthenticator {
	return &JWTAuthenticator{
		secret: []byte(jwtSecret),
		aud:    aud,
		iss:    iss,
	}
}

// GenerateToken creates a signed JWT string using the provided custom claims and the configured secret.
func (a *JWTAuthenticator) GenerateToken(claims CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(a.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %w", err)
	}

	return signedToken, nil
}

// ValidateToken parses, verifies, and extracts custom claims from a given token string.
func (a *JWTAuthenticator) ValidateToken(tokenString string) (*CustomClaims, error) {
	claims := &CustomClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.Join(ErrUnexpectedSigningMethod)
			}

			return a.secret, nil
		},
		jwt.WithAudience(a.aud),
		jwt.WithIssuer(a.iss),
		jwt.WithExpirationRequired(),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse or verify JWT token: %w", err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
