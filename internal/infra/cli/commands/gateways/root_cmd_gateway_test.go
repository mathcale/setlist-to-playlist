package gateways

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/setlist-to-playlist/internal/entities/music"
	"github.com/mathcale/setlist-to-playlist/internal/entities/setlistfm"
	"github.com/mathcale/setlist-to-playlist/internal/tests/mocks"
)

type RootCmdGatewayTestSuite struct {
	suite.Suite
	LoggerMock                *mocks.LoggerMock
	AuthenticatorMock         *mocks.AuthenticatorMock
	GetSetlistByIDUseCaseMock *mocks.SetlistFMGetSetlistByIDUseCaseMock
	FetchSongsUseCaseMock     *mocks.FetchSongsUseCaseMock
	CreatePlaylistUseCaseMock *mocks.CreatePlaylistUseCaseMock
	AddTracksUseCaseMock      *mocks.AddTracksUseCaseMock

	Gateway RootCmdGatewayInterface
}

func (s *RootCmdGatewayTestSuite) SetupTest() {
	s.LoggerMock = new(mocks.LoggerMock)
	s.AuthenticatorMock = new(mocks.AuthenticatorMock)
	s.GetSetlistByIDUseCaseMock = new(mocks.SetlistFMGetSetlistByIDUseCaseMock)
	s.FetchSongsUseCaseMock = new(mocks.FetchSongsUseCaseMock)
	s.CreatePlaylistUseCaseMock = new(mocks.CreatePlaylistUseCaseMock)
	s.AddTracksUseCaseMock = new(mocks.AddTracksUseCaseMock)

	s.Gateway = NewRootCmdGateway(
		s.LoggerMock,
		s.AuthenticatorMock,
		s.GetSetlistByIDUseCaseMock,
		s.FetchSongsUseCaseMock,
		s.CreatePlaylistUseCaseMock,
		s.AddTracksUseCaseMock,
	)
}

func (s *RootCmdGatewayTestSuite) cleanMocks() {
	s.LoggerMock.ExpectedCalls = nil
	s.LoggerMock.Calls = nil
	s.AuthenticatorMock.ExpectedCalls = nil
	s.AuthenticatorMock.Calls = nil
	s.GetSetlistByIDUseCaseMock.ExpectedCalls = nil
	s.GetSetlistByIDUseCaseMock.Calls = nil
	s.FetchSongsUseCaseMock.ExpectedCalls = nil
	s.FetchSongsUseCaseMock.Calls = nil
	s.CreatePlaylistUseCaseMock.ExpectedCalls = nil
	s.CreatePlaylistUseCaseMock.Calls = nil
	s.AddTracksUseCaseMock.ExpectedCalls = nil
	s.AddTracksUseCaseMock.Calls = nil
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

func (s *RootCmdGatewayTestSuite) TestAuthenticate() {
	s.Run("Should delegate to the Authenticator", func() {
		defer s.cleanMocks()

		s.AuthenticatorMock.On("Authenticate", mock.Anything).Return(nil)

		err := s.Gateway.Authenticate(context.Background())

		s.NoError(err)
		s.AuthenticatorMock.AssertNumberOfCalls(s.T(), "Authenticate", 1)
	})

	s.Run("Should return an error when authentication fails", func() {
		defer s.cleanMocks()

		s.AuthenticatorMock.On("Authenticate", mock.Anything).Return(errors.New("any-error"))

		err := s.Gateway.Authenticate(context.Background())

		s.Error(err)
	})
}

func (s *RootCmdGatewayTestSuite) TestFetchSongs() {
	s.Run("Should fetch songs on Spotify", func() {
		defer s.cleanMocks()

		expected := &music.FetchSongsOutput{
			Artist: "any-artist",
			Songs: []music.Song{
				{ID: "any-song-id-1", Title: "any-song-1", Album: "any-album-1"},
				{ID: "any-song-id-2", Title: "any-song-2", Album: "any-album-1"},
				{ID: "any-song-id-3", Title: "any-song-3", Album: "any-album-1"},
			},
		}

		songs := []string{"any-song-1", "any-song-2", "any-song-3"}

		s.FetchSongsUseCaseMock.
			On("Execute", mock.Anything, mock.Anything, mock.Anything).
			Return(expected, nil)

		result, err := s.Gateway.FetchSongs(context.Background(), songs, "any-artist")

		s.NoError(err)
		s.Equal(expected, result)
	})

	s.Run("Should return an error when failing to fetch songs from Spotify", func() {
		defer s.cleanMocks()

		songs := []string{"any-song-1", "any-song-2", "any-song-3"}

		s.FetchSongsUseCaseMock.
			On("Execute", mock.Anything, mock.Anything, mock.Anything).
			Return(nil, errors.New("any-error"))

		_, err := s.Gateway.FetchSongs(context.Background(), songs, "any-artist")

		s.Error(err)
	})
}

func (s *RootCmdGatewayTestSuite) TestCreatePlaylist() {
	s.Run("Should create a playlist on Spotify", func() {
		defer s.cleanMocks()

		ctx := context.Background()

		expected := &music.Playlist{
			ID:  "any-playlist-id",
			URL: "https://open.spotify.com/playlist/any-playlist-id",
		}

		songs := []music.Song{
			{ID: "any-song-id-1", Title: "any-song-1", Album: "any-album-1"},
			{ID: "any-song-id-2", Title: "any-song-2", Album: "any-album-1"},
			{ID: "any-song-id-3", Title: "any-song-3", Album: "any-album-1"},
		}

		s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return(nil)
		s.CreatePlaylistUseCaseMock.
			On("Execute", mock.Anything, mock.Anything).
			Return(expected, nil)

		s.AddTracksUseCaseMock.
			On("Execute", mock.Anything, mock.Anything).
			Return(nil)

		result, err := s.Gateway.CreatePlaylist(ctx, "any-playlist-name", songs)

		s.NoError(err)
		s.Equal(expected.URL, *result)
	})

	s.Run("Should return an error when failing to create a playlist on Spotify", func() {
		defer s.cleanMocks()

		ctx := context.Background()

		songs := []music.Song{
			{ID: "any-song-id-1", Title: "any-song-1", Album: "any-album-1"},
			{ID: "any-song-id-2", Title: "any-song-2", Album: "any-album-1"},
			{ID: "any-song-id-3", Title: "any-song-3", Album: "any-album-1"},
		}

		s.CreatePlaylistUseCaseMock.
			On("Execute", mock.Anything, mock.Anything).
			Return(nil, errors.New("any-error"))

		_, err := s.Gateway.CreatePlaylist(ctx, "any-playlist-name", songs)

		s.Error(err)
		s.ErrorContains(err, "any-error")
	})

	s.Run("Should return an error when failing to add tracks to a playlist on Spotify", func() {
		defer s.cleanMocks()

		ctx := context.Background()

		expected := &music.Playlist{
			ID:  "any-playlist-id",
			URL: "https://open.spotify.com/playlist/any-playlist-id",
		}

		songs := []music.Song{
			{ID: "any-song-id-1", Title: "any-song-1", Album: "any-album-1"},
			{ID: "any-song-id-2", Title: "any-song-2", Album: "any-album-1"},
			{ID: "any-song-id-3", Title: "any-song-3", Album: "any-album-1"},
		}

		s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return(nil)
		s.CreatePlaylistUseCaseMock.
			On("Execute", mock.Anything, mock.Anything).
			Return(expected, nil)

		s.AddTracksUseCaseMock.
			On("Execute", mock.Anything, mock.Anything).
			Return(errors.New("any-error"))

		_, err := s.Gateway.CreatePlaylist(ctx, "any-playlist-name", songs)

		s.Error(err)
		s.ErrorContains(err, "any-error")
	})
}
