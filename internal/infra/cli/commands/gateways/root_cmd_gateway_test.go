package gateways

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	spotifyclient "github.com/mathcale/setlist-to-playlist/internal/clients/spotify"
	"github.com/mathcale/setlist-to-playlist/internal/entities/setlistfm"
	spotifyentities "github.com/mathcale/setlist-to-playlist/internal/entities/spotify"
	oauth2util "github.com/mathcale/setlist-to-playlist/internal/pkg/oauth2"
	"github.com/mathcale/setlist-to-playlist/internal/tests/mocks"
)

type RootCmdGatewayTestSuite struct {
	suite.Suite
	LoggerMock                            *mocks.LoggerMock
	WebServerMock                         *mocks.WebServerMock
	SpotifyClientMock                     *mocks.SpotifyClientMock
	GetSetlistByIDUseCaseMock             *mocks.SetlistFMGetSetlistByIDUseCaseMock
	FetchSongsOnSpotifyUseCaseMock        *mocks.FetchSongsOnSpotifyUseCaseMock
	CreatePlaylistOnSpotifyUseCaseMock    *mocks.CreatePlaylistOnSpotifyUseCaseMock
	AddTracksToSpotifyPlaylistUseCaseMock *mocks.AddTracksToSpotifyPlaylistUseCaseMock
	SpotifyUserAuthenticationUseCaseMock  *mocks.SpotifyUserAuthenticationUseCaseMock
	GeneratedPKCECodes                    oauth2util.GenerateOutput
	State                                 string
	SpotifyClientChannel                  chan spotifyclient.AuthenticatedClient

	Gateway RootCmdGatewayInterface
}

func (s *RootCmdGatewayTestSuite) SetupTest() {
	s.LoggerMock = new(mocks.LoggerMock)
	s.WebServerMock = new(mocks.WebServerMock)
	s.SpotifyClientMock = new(mocks.SpotifyClientMock)
	s.GetSetlistByIDUseCaseMock = new(mocks.SetlistFMGetSetlistByIDUseCaseMock)
	s.FetchSongsOnSpotifyUseCaseMock = new(mocks.FetchSongsOnSpotifyUseCaseMock)
	s.CreatePlaylistOnSpotifyUseCaseMock = new(mocks.CreatePlaylistOnSpotifyUseCaseMock)
	s.AddTracksToSpotifyPlaylistUseCaseMock = new(mocks.AddTracksToSpotifyPlaylistUseCaseMock)
	s.SpotifyUserAuthenticationUseCaseMock = new(mocks.SpotifyUserAuthenticationUseCaseMock)
	s.GeneratedPKCECodes = oauth2util.GenerateOutput{
		CodeChallenge: "any-code-challenge",
		CodeVerifier:  "any-code-verifier",
	}
	s.State = "any-state"
	s.SpotifyClientChannel = make(chan spotifyclient.AuthenticatedClient)

	s.Gateway = NewRootCmdGateway(
		s.LoggerMock,
		s.WebServerMock,
		s.SpotifyClientMock,
		s.GetSetlistByIDUseCaseMock,
		s.FetchSongsOnSpotifyUseCaseMock,
		s.CreatePlaylistOnSpotifyUseCaseMock,
		s.AddTracksToSpotifyPlaylistUseCaseMock,
		s.SpotifyUserAuthenticationUseCaseMock,
		s.GeneratedPKCECodes,
		s.State,
		s.SpotifyClientChannel,
	)
}

func (s *RootCmdGatewayTestSuite) cleanMocks() {
	s.LoggerMock.ExpectedCalls = nil
	s.LoggerMock.Calls = nil
	s.WebServerMock.ExpectedCalls = nil
	s.WebServerMock.Calls = nil
	s.SpotifyClientMock.ExpectedCalls = nil
	s.SpotifyClientMock.Calls = nil
	s.GetSetlistByIDUseCaseMock.ExpectedCalls = nil
	s.GetSetlistByIDUseCaseMock.Calls = nil
	s.FetchSongsOnSpotifyUseCaseMock.ExpectedCalls = nil
	s.FetchSongsOnSpotifyUseCaseMock.Calls = nil
	s.CreatePlaylistOnSpotifyUseCaseMock.ExpectedCalls = nil
	s.CreatePlaylistOnSpotifyUseCaseMock.Calls = nil
	s.AddTracksToSpotifyPlaylistUseCaseMock.ExpectedCalls = nil
	s.AddTracksToSpotifyPlaylistUseCaseMock.Calls = nil
	s.SpotifyUserAuthenticationUseCaseMock.ExpectedCalls = nil
	s.SpotifyUserAuthenticationUseCaseMock.Calls = nil
}

func TestRootCmdGateway(t *testing.T) {
	suite.Run(t, new(RootCmdGatewayTestSuite))
}

func (s *RootCmdGatewayTestSuite) TestGetTracksFromSetlist() {
	s.Run("Should return a set of tracks from a Setlist.fm URL", func() {
		defer s.cleanMocks()

		expected := &setlistfm.Set{
			ID: "any-set-id",
			Sets: setlistfm.Sets{
				Set: []setlistfm.Songs{
					{
						Song: []setlistfm.Song{
							{Name: "any-song-1"},
							{Name: "any-song-2"},
							{Name: "any-song-3"},
						},
					},
					{
						Song: []setlistfm.Song{
							{Name: "any-song-4"},
						},
						Encore: 1,
					},
				},
			},
		}

		s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return(nil)
		s.GetSetlistByIDUseCaseMock.On("Execute", mock.Anything).Return(expected, nil)

		result, err := s.Gateway.GetTracksFromSetlist("https://www.setlist.fm/setlist/blink182/2024/autodromo-de-interlagos-sao-paulo-brazil-53aa1325.html")

		s.NoError(err)
		s.Equal(expected, result)
	})

	s.Run("Should return an error when failing to extract Setlist.fm ID from URL", func() {
		defer s.cleanMocks()

		s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return(nil)
		s.GetSetlistByIDUseCaseMock.
			On("Execute", mock.Anything).
			Return(nil, errors.New("any-validation-error"))

		_, err := s.Gateway.GetTracksFromSetlist("any-setlistfm-url")

		s.Error(err)
	})

	s.Run("Should return an error when failing to get Setlist.fm data", func() {
		defer s.cleanMocks()

		s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return(nil)
		s.GetSetlistByIDUseCaseMock.
			On("Execute", mock.Anything).
			Return(nil, errors.New("any-error"))

		_, err := s.Gateway.GetTracksFromSetlist("any-setlistfm-url")

		s.Error(err)
	})
}

func (s *RootCmdGatewayTestSuite) TestStartWebServer() {
	s.Run("Should start the web server", func() {
		defer s.cleanMocks()

		s.WebServerMock.On("Start").Return()

		s.Gateway.StartWebServer()

		s.WebServerMock.AssertCalled(s.T(), "Start")
	})
}

func (s *RootCmdGatewayTestSuite) TestHandleSpotifyAuthentication() {
	s.Run("Should handle Spotify authentication", func() {
		defer s.cleanMocks()

		s.SpotifyUserAuthenticationUseCaseMock.
			On("Execute", mock.Anything, s.GeneratedPKCECodes, s.State).
			Return(nil)

		err := s.Gateway.HandleSpotifyAuthentication(context.Background())

		s.NoError(err)
	})

	s.Run("Should return an error when failing to authenticate on Spotify", func() {
		defer s.cleanMocks()

		s.SpotifyUserAuthenticationUseCaseMock.
			On("Execute", mock.Anything, s.GeneratedPKCECodes, s.State).
			Return(errors.New("any-error"))

		err := s.Gateway.HandleSpotifyAuthentication(context.Background())

		s.Error(err)
	})
}

func (s *RootCmdGatewayTestSuite) TestFetchSongsOnSpotify() {
	s.Run("Should fetch songs on Spotify", func() {
		defer s.cleanMocks()

		expected := &spotifyentities.FindAllSongsOutput{
			Artist: "any-artist",
			Songs: []spotifyentities.Song{
				{ID: "any-song-id-1", Title: "any-song-1", Album: "any-album-1"},
				{ID: "any-song-id-2", Title: "any-song-2", Album: "any-album-1"},
				{ID: "any-song-id-3", Title: "any-song-3", Album: "any-album-1"},
			},
		}

		songs := []string{"any-song-1", "any-song-2", "any-song-3"}

		s.FetchSongsOnSpotifyUseCaseMock.
			On("Execute", mock.Anything, mock.Anything, mock.Anything).
			Return(expected, nil)

		result, err := s.Gateway.FetchSongsOnSpotify(context.Background(), songs, "any-artist")

		s.NoError(err)
		s.Equal(expected, result)
	})

	s.Run("Should return an error when failing to fetch songs from Spotify", func() {
		defer s.cleanMocks()

		songs := []string{"any-song-1", "any-song-2", "any-song-3"}

		s.FetchSongsOnSpotifyUseCaseMock.
			On("Execute", mock.Anything, mock.Anything, mock.Anything).
			Return(nil, errors.New("any-error"))

		_, err := s.Gateway.FetchSongsOnSpotify(context.Background(), songs, "any-artist")

		s.Error(err)
	})
}

func (s *RootCmdGatewayTestSuite) TestCreatePlaylistOnSpotify() {
	s.Run("Should create a playlist on Spotify", func() {
		defer s.cleanMocks()

		ctx := context.Background()

		expected := &spotifyentities.CreatePlaylistOutput{
			ID:  "any-playlist-id",
			URL: "https://open.spotify.com/playlist/any-playlist-id",
		}

		songs := []spotifyentities.Song{
			{ID: "any-song-id-1", Title: "any-song-1", Album: "any-album-1"},
			{ID: "any-song-id-2", Title: "any-song-2", Album: "any-album-1"},
			{ID: "any-song-id-3", Title: "any-song-3", Album: "any-album-1"},
		}

		s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return(nil)
		s.CreatePlaylistOnSpotifyUseCaseMock.
			On("Execute", mock.Anything, mock.Anything).
			Return(expected, nil)

		s.AddTracksToSpotifyPlaylistUseCaseMock.
			On("Execute", mock.Anything, mock.Anything).
			Return(nil)

		result, err := s.Gateway.CreatePlaylistOnSpotify(ctx, "any-playlist-name", songs)

		s.NoError(err)
		s.Equal(expected.URL, *result)
	})

	s.Run("Should return an error when failing to create a playlist on Spotify", func() {
		defer s.cleanMocks()

		ctx := context.Background()

		songs := []spotifyentities.Song{
			{ID: "any-song-id-1", Title: "any-song-1", Album: "any-album-1"},
			{ID: "any-song-id-2", Title: "any-song-2", Album: "any-album-1"},
			{ID: "any-song-id-3", Title: "any-song-3", Album: "any-album-1"},
		}

		s.CreatePlaylistOnSpotifyUseCaseMock.
			On("Execute", mock.Anything, mock.Anything).
			Return(nil, errors.New("any-error"))

		_, err := s.Gateway.CreatePlaylistOnSpotify(ctx, "any-playlist-name", songs)

		s.Error(err)
		s.ErrorContains(err, "any-error")
	})

	s.Run("Should return an error when failing to add tracks to a playlist on Spotify", func() {
		defer s.cleanMocks()

		ctx := context.Background()

		expected := &spotifyentities.CreatePlaylistOutput{
			ID:  "any-playlist-id",
			URL: "https://open.spotify.com/playlist/any-playlist-id",
		}

		songs := []spotifyentities.Song{
			{ID: "any-song-id-1", Title: "any-song-1", Album: "any-album-1"},
			{ID: "any-song-id-2", Title: "any-song-2", Album: "any-album-1"},
			{ID: "any-song-id-3", Title: "any-song-3", Album: "any-album-1"},
		}

		s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return(nil)
		s.CreatePlaylistOnSpotifyUseCaseMock.
			On("Execute", mock.Anything, mock.Anything).
			Return(expected, nil)

		s.AddTracksToSpotifyPlaylistUseCaseMock.
			On("Execute", mock.Anything, mock.Anything).
			Return(errors.New("any-error"))

		_, err := s.Gateway.CreatePlaylistOnSpotify(ctx, "any-playlist-name", songs)

		s.Error(err)
		s.ErrorContains(err, "any-error")
	})
}
