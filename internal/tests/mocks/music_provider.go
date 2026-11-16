package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/mathcale/setlist-to-playlist/internal/entities/music"
)

type MusicProviderMock struct {
	mock.Mock
}

func (m *MusicProviderMock) SearchCandidates(ctx context.Context, title string, artist string) ([]music.Candidate, error) {
	args := m.Called(ctx, title, artist)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]music.Candidate), args.Error(1)
}

func (m *MusicProviderMock) CreatePlaylist(ctx context.Context, title string, description string) (*music.Playlist, error) {
	args := m.Called(ctx, title, description)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*music.Playlist), args.Error(1)
}

func (m *MusicProviderMock) AddTracks(ctx context.Context, playlistID string, trackIDs []string) error {
	args := m.Called(ctx, playlistID, trackIDs)
	return args.Error(0)
}
