// Package auth provides authentication mechanisms, token generation, and validation interfaces.
package auth

// Authenticator defines methods for generating and validating authentication tokens.
type Authenticator interface {
	// GenerateToken creates a new token based on the provided claims.
	GenerateToken(claims CustomClaims) (string, error)

	// ValidateToken verifies the token string and returns the associated claims if valid.
	ValidateToken(tokenString string) (*CustomClaims, error)
}
