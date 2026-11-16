package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// AuthenticatorMock is a fake music.Authenticator.
type AuthenticatorMock struct {
	mock.Mock
}

func (m *AuthenticatorMock) Authenticate(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *AuthenticatorMock) Reauthenticate(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// CredentialsEnsurerMock is a fake spotify.CredentialsEnsurer.
type CredentialsEnsurerMock struct {
	mock.Mock
}

func (m *CredentialsEnsurerMock) Ensure() error {
	args := m.Called()
	return args.Error(0)
}
