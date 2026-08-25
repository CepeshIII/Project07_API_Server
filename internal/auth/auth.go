package auth

type Authenticator interface {
	GenerateToken(claims CustomClaims) (string, error)
	ValidateToken(tokenString string) (*CustomClaims, error)
}
