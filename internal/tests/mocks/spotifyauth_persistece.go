package mocks

import (
	"github.com/stretchr/testify/mock"

	"github.com/mathcale/setlist-to-playlist/internal/entities/spotify"
)

type SpotifyAuthPersistenceMock struct {
	mock.Mock
}

func (m *SpotifyAuthPersistenceMock) Read() (*spotify.SpotifyUserAuthData, error) {
	args := m.Called()

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*spotify.SpotifyUserAuthData), args.Error(1)
}

func (m *SpotifyAuthPersistenceMock) Write(data spotify.SpotifyUserAuthData) error {
	args := m.Called(data)
	return args.Error(0)
}
