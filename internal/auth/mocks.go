package auth

import (
	"github.com/stretchr/testify/mock"
)

type MockAuthenticator struct {
	mock.Mock
}

func (m *MockAuthenticator) GenerateToken(claims CustomClaims) (string, error) {
	args := m.Called(claims)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}

	token, ok := args.Get(0).(string)
	if !ok {
		return "", args.Error(1)
	}

	return token, args.Error(1)
}

func (m *MockAuthenticator) ValidateToken(tokenString string) (*CustomClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	claims, ok := args.Get(0).(*CustomClaims)
	if !ok {
		return nil, args.Error(1)
	}

	return claims, args.Error(1)
}
