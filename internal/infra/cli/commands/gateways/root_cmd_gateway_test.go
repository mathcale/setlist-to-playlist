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
	ProviderMock              *mocks.MusicProviderMock
	GetSetlistByIDUseCaseMock *mocks.SetlistFMGetSetlistByIDUseCaseMock

	Gateway RootCmdGatewayInterface
}

func (s *RootCmdGatewayTestSuite) SetupTest() {
	s.LoggerMock = new(mocks.LoggerMock)
	s.AuthenticatorMock = new(mocks.AuthenticatorMock)
	s.ProviderMock = new(mocks.MusicProviderMock)
	s.GetSetlistByIDUseCaseMock = new(mocks.SetlistFMGetSetlistByIDUseCaseMock)

	registry := music.NewProviderRegistry().
		Register(music.ProviderSpotify, s.ProviderMock, s.AuthenticatorMock)

	s.Gateway = NewRootCmdGateway(s.LoggerMock, registry, s.GetSetlistByIDUseCaseMock)
}

func (s *RootCmdGatewayTestSuite) cleanMocks() {
	s.LoggerMock.ExpectedCalls = nil
	s.LoggerMock.Calls = nil
	s.AuthenticatorMock.ExpectedCalls = nil
	s.AuthenticatorMock.Calls = nil
	s.ProviderMock.ExpectedCalls = nil
	s.ProviderMock.Calls = nil
	s.GetSetlistByIDUseCaseMock.ExpectedCalls = nil
	s.GetSetlistByIDUseCaseMock.Calls = nil
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

func (s *RootCmdGatewayTestSuite) TestCheckProvider() {
	s.Run("Should accept a registered provider", func() {
		s.NoError(s.Gateway.CheckProvider(music.ProviderSpotify))
	})

	s.Run("Should return ErrProviderUnavailable for an unregistered provider", func() {
		err := s.Gateway.CheckProvider(music.ProviderAppleMusic)

		s.ErrorIs(err, music.ErrProviderUnavailable)
		s.ErrorContains(err, "Apple Music")
	})
}

func (s *RootCmdGatewayTestSuite) TestAuthenticate() {
	s.Run("Should delegate to the registered provider's Authenticator", func() {
		defer s.cleanMocks()

		s.AuthenticatorMock.On("Authenticate", mock.Anything).Return(nil)

		s.NoError(s.Gateway.Authenticate(context.Background(), music.ProviderSpotify))
		s.AuthenticatorMock.AssertNumberOfCalls(s.T(), "Authenticate", 1)
	})

	s.Run("Should return an error when authentication fails", func() {
		defer s.cleanMocks()

		s.AuthenticatorMock.On("Authenticate", mock.Anything).Return(errors.New("any-error"))

		s.Error(s.Gateway.Authenticate(context.Background(), music.ProviderSpotify))
	})

	s.Run("Should not authenticate an unregistered provider", func() {
		defer s.cleanMocks()

		err := s.Gateway.Authenticate(context.Background(), music.ProviderAppleMusic)

		s.ErrorIs(err, music.ErrProviderUnavailable)
		s.AuthenticatorMock.AssertNotCalled(s.T(), "Authenticate", mock.Anything)
	})
}

func (s *RootCmdGatewayTestSuite) TestFetchSongs() {
	s.Run("Should search on the chosen provider and return the best matches", func() {
		defer s.cleanMocks()

		s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return()
		s.ProviderMock.
			On("SearchCandidates", mock.Anything, "any-song-1", "any-artist").
			Return([]music.Candidate{{
				TrackID: "id-1", Title: "any-song-1", Artist: "any-artist",
				Album: "any-album", AlbumType: music.AlbumTypeAlbum,
			}}, nil)

		result, err := s.Gateway.FetchSongs(
			context.Background(), music.ProviderSpotify, []string{"any-song-1"}, "any-artist",
		)

		s.NoError(err)
		s.Equal([]music.Song{{ID: "id-1", Title: "any-song-1", Album: "any-album"}}, result.Songs)
	})

	s.Run("Should return an error when the provider search fails", func() {
		defer s.cleanMocks()

		s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return()
		s.ProviderMock.
			On("SearchCandidates", mock.Anything, mock.Anything, mock.Anything).
			Return(nil, errors.New("any-error"))

		_, err := s.Gateway.FetchSongs(
			context.Background(), music.ProviderSpotify, []string{"any-song-1"}, "any-artist",
		)

		s.Error(err)
	})

	s.Run("Should return ErrProviderUnavailable for an unregistered provider", func() {
		_, err := s.Gateway.FetchSongs(
			context.Background(), music.ProviderAppleMusic, []string{"any-song-1"}, "any-artist",
		)

		s.ErrorIs(err, music.ErrProviderUnavailable)
	})
}

func (s *RootCmdGatewayTestSuite) TestCreatePlaylist() {
	songs := []music.Song{
		{ID: "any-song-id-1", Title: "any-song-1", Album: "any-album-1"},
		{ID: "any-song-id-2", Title: "any-song-2", Album: "any-album-1"},
	}

	s.Run("Should create a playlist and add the tracks on the chosen provider", func() {
		defer s.cleanMocks()

		s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return()
		s.ProviderMock.
			On("CreatePlaylist", mock.Anything, "any-playlist-name", mock.Anything).
			Return(&music.Playlist{ID: "any-playlist-id", URL: "any-url"}, nil)
		s.ProviderMock.
			On("AddTracks", mock.Anything, "any-playlist-id", []string{"any-song-id-1", "any-song-id-2"}).
			Return(nil)

		result, err := s.Gateway.CreatePlaylist(
			context.Background(), music.ProviderSpotify, "any-playlist-name", songs,
		)

		s.NoError(err)
		s.Equal("any-url", *result)
	})

	s.Run("Should return an error when failing to create a playlist", func() {
		defer s.cleanMocks()

		s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return()
		s.ProviderMock.
			On("CreatePlaylist", mock.Anything, mock.Anything, mock.Anything).
			Return(nil, errors.New("any-error"))

		_, err := s.Gateway.CreatePlaylist(context.Background(), music.ProviderSpotify, "n", songs)

		s.ErrorContains(err, "any-error")
	})

	s.Run("Should return an error when failing to add tracks", func() {
		defer s.cleanMocks()

		s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return()
		s.ProviderMock.
			On("CreatePlaylist", mock.Anything, mock.Anything, mock.Anything).
			Return(&music.Playlist{ID: "id", URL: "url"}, nil)
		s.ProviderMock.
			On("AddTracks", mock.Anything, mock.Anything, mock.Anything).
			Return(errors.New("any-error"))

		_, err := s.Gateway.CreatePlaylist(context.Background(), music.ProviderSpotify, "n", songs)

		s.ErrorContains(err, "any-error")
	})

	s.Run("Should return ErrProviderUnavailable for an unregistered provider", func() {
		_, err := s.Gateway.CreatePlaylist(context.Background(), music.ProviderAppleMusic, "n", songs)

		s.ErrorIs(err, music.ErrProviderUnavailable)
	})
}
