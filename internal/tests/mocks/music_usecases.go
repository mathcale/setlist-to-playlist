package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/mathcale/setlist-to-playlist/internal/entities/music"
)

type FetchSongsUseCaseMock struct {
	mock.Mock
}

type CreatePlaylistUseCaseMock struct {
	mock.Mock
}

type AddTracksUseCaseMock struct {
	mock.Mock
}

func (m *FetchSongsUseCaseMock) Execute(ctx context.Context, songs []string, artist string) (*music.FetchSongsOutput, error) {
	args := m.Called(ctx, songs, artist)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*music.FetchSongsOutput), args.Error(1)
}

func (m *CreatePlaylistUseCaseMock) Execute(ctx context.Context, input music.CreatePlaylistInput) (*music.Playlist, error) {
	args := m.Called(ctx, input)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*music.Playlist), args.Error(1)
}

func (m *AddTracksUseCaseMock) Execute(ctx context.Context, input music.AddTracksInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}
