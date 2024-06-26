package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	entities "github.com/mathcale/setlist-to-playlist/internal/entities/spotify"
	oauth2util "github.com/mathcale/setlist-to-playlist/internal/pkg/oauth2"
)

type SpotifyUserAuthenticationUseCaseGatewayMock struct {
	mock.Mock
}

func (m *SpotifyUserAuthenticationUseCaseGatewayMock) ValidatePersistedToken() (*entities.SpotifyUserAuthData, error) {
	args := m.Called()

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*entities.SpotifyUserAuthData), args.Error(1)
}

func (m *SpotifyUserAuthenticationUseCaseGatewayMock) AuthenticateUser(state string, pkceCodes oauth2util.GenerateOutput) error {
	args := m.Called(state, pkceCodes)
	return args.Error(0)
}

func (m *SpotifyUserAuthenticationUseCaseGatewayMock) RefreshToken(ctx context.Context, authData *entities.SpotifyUserAuthData) error {
	args := m.Called(ctx, authData)
	return args.Error(0)
}
