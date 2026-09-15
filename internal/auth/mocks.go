package auth

import "github.com/stretchr/testify/mock"

type MockAuthenticator struct {
	mock.Mock
}

func (m *MockAuthenticator) GenerateToken(claims CustomClaims) (string, error) {
	args := m.Called(claims)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.Get(0).(string), args.Error(1)
}

func (m *MockAuthenticator) ValidateToken(tokenString string) (*CustomClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CustomClaims), args.Error(1)
}
