package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/mathcale/setlist-to-playlist/internal/entities/music"
	"github.com/mathcale/setlist-to-playlist/internal/entities/setlistfm"
)

type RootCmdGatewayMock struct {
	mock.Mock
}

func (m *RootCmdGatewayMock) GetTracksFromSetlist(setlistfmURL string) (*setlistfm.Set, error) {
	args := m.Called(setlistfmURL)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*setlistfm.Set), args.Error(1)
}

func (m *RootCmdGatewayMock) CheckProvider(id music.ProviderID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *RootCmdGatewayMock) Authenticate(ctx context.Context, id music.ProviderID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *RootCmdGatewayMock) FetchSongs(
	ctx context.Context,
	id music.ProviderID,
	songTitles []string,
	artist string,
) (*music.FetchSongsOutput, error) {
	args := m.Called(ctx, id, songTitles, artist)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*music.FetchSongsOutput), args.Error(1)
}

func (m *RootCmdGatewayMock) CreatePlaylist(
	ctx context.Context,
	id music.ProviderID,
	playlistName string,
	songs []music.Song,
) (*string, error) {
	args := m.Called(ctx, id, playlistName, songs)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*string), args.Error(1)
}
