package mocks

import (
	"context"
	"net/http"

	"github.com/stretchr/testify/mock"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"

	client "github.com/mathcale/setlist-to-playlist/internal/clients/spotify"
	entities "github.com/mathcale/setlist-to-playlist/internal/entities/spotify"
	oauth2util "github.com/mathcale/setlist-to-playlist/internal/pkg/oauth2"
)

type SpotifyClientMock struct {
	mock.Mock
}

func (m *SpotifyClientMock) GetToken(
	ctx context.Context,
	r *http.Request,
	state string,
	genCodes oauth2util.GenerateOutput,
) (*oauth2.Token, error) {
	args := m.Called(ctx, r, state, genCodes)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*oauth2.Token), args.Error(1)
}

func (m *SpotifyClientMock) GetAuthURL(state string, genCodes oauth2util.GenerateOutput) string {
	args := m.Called(state, genCodes)
	return args.String(0)
}

func (m *SpotifyClientMock) NewAPIClient(ctx context.Context, tok *oauth2.Token) *spotify.Client {
	args := m.Called(ctx, tok)
	return args.Get(0).(*spotify.Client)
}

func (m *SpotifyClientMock) SetAuthenticatedClient(ch chan client.AuthenticatedClient) {
	m.Called(ch)
}

func (m *SpotifyClientMock) SetAuthenticatedClientFromInstance(client client.AuthenticatedClient) {
	m.Called(client)
}

func (m *SpotifyClientMock) CurrentUser(ctx context.Context) (*spotify.PrivateUser, error) {
	args := m.Called(ctx)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*spotify.PrivateUser), args.Error(1)
}

func (m *SpotifyClientMock) CurrentSession() (*oauth2.Token, error) {
	args := m.Called()

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*oauth2.Token), args.Error(1)
}

func (m *SpotifyClientMock) RefreshToken(
	ctx context.Context,
	tok *oauth2.Token,
) (*oauth2.Token, error) {
	args := m.Called(ctx, tok)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*oauth2.Token), args.Error(1)
}

func (m *SpotifyClientMock) FindAllSongsByName(
	ctx context.Context,
	name []string,
	artist string,
) (*entities.FindAllSongsOutput, error) {
	args := m.Called(ctx, name, artist)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*entities.FindAllSongsOutput), args.Error(1)
}

func (m *SpotifyClientMock) CreatePlaylist(
	ctx context.Context,
	title string,
	description string,
) (*entities.CreatePlaylistOutput, error) {
	args := m.Called(ctx, title, description)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*entities.CreatePlaylistOutput), args.Error(1)
}

func (m *SpotifyClientMock) AddTracksToPlaylist(
	ctx context.Context,
	input entities.AddTracksToPlaylistClientInput,
) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}
