package mocks

import (
	"context"
	"net/http"

	"github.com/stretchr/testify/mock"
	"github.com/zmb3/spotify/v2"

	entities "github.com/mathcale/setlist-to-playlist/internal/entities/spotify"
	oauth2util "github.com/mathcale/setlist-to-playlist/internal/pkg/oauth2"
)

type SpotifyAuthCallbackUseCaseMock struct {
	mock.Mock
}

type FetchSongsOnSpotifyUseCaseMock struct {
	mock.Mock
}

type CreatePlaylistOnSpotifyUseCaseMock struct {
	mock.Mock
}

type AddTracksToSpotifyPlaylistUseCaseMock struct {
	mock.Mock
}

type SpotifyUserAuthenticationUseCaseMock struct {
	mock.Mock
}

func (m *SpotifyAuthCallbackUseCaseMock) Execute(
	ctx context.Context,
	r *http.Request,
	state string,
	genCodes oauth2util.GenerateOutput,
) (*spotify.Client, error) {
	args := m.Called(ctx, r, state, genCodes)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*spotify.Client), args.Error(1)
}

func (m *FetchSongsOnSpotifyUseCaseMock) Execute(
	ctx context.Context,
	songs []string,
	artist string,
) (*entities.FindAllSongsOutput, error) {
	args := m.Called(ctx, songs, artist)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*entities.FindAllSongsOutput), args.Error(1)
}

func (m *CreatePlaylistOnSpotifyUseCaseMock) Execute(
	ctx context.Context,
	input entities.CreatePlaylistInput,
) (*entities.CreatePlaylistOutput, error) {
	args := m.Called(ctx, input)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*entities.CreatePlaylistOutput), args.Error(1)
}

func (m *AddTracksToSpotifyPlaylistUseCaseMock) Execute(
	ctx context.Context,
	input entities.AddTracksToPlaylistInput,
) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}

func (m *SpotifyUserAuthenticationUseCaseMock) Execute(
	ctx context.Context,
	pkceCodes oauth2util.GenerateOutput,
	state string,
) error {
	args := m.Called(ctx, pkceCodes, state)
	return args.Error(0)
}
