package spotify

import (
	"context"
	"errors"
	"testing"

	entities "github.com/mathcale/setlist-to-playlist/internal/entities/spotify"
	"github.com/mathcale/setlist-to-playlist/internal/tests/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type AddTracksToPlaylistUseCaseTestSuite struct {
	suite.Suite

	ClientMock *mocks.SpotifyClientMock
	LoggerMock *mocks.LoggerMock

	UseCase AddTracksToPlaylistUseCaseInterface
}

func (s *AddTracksToPlaylistUseCaseTestSuite) SetupTest() {
	s.ClientMock = new(mocks.SpotifyClientMock)
	s.LoggerMock = new(mocks.LoggerMock)

	s.UseCase = NewAddTracksToPlaylistUseCase(
		s.ClientMock,
		s.LoggerMock,
	)
}

func (s *AddTracksToPlaylistUseCaseTestSuite) cleanMocks() {
	s.ClientMock.ExpectedCalls = nil
	s.ClientMock.Calls = nil
	s.LoggerMock.ExpectedCalls = nil
	s.LoggerMock.Calls = nil
}

func TestAddTracksToPlaylistUseCase(t *testing.T) {
	suite.Run(t, new(AddTracksToPlaylistUseCaseTestSuite))
}

func (s *AddTracksToPlaylistUseCaseTestSuite) TestExecute() {
	s.Run("should add tracks to playlist", func() {
		defer s.cleanMocks()

		input := entities.AddTracksToPlaylistInput{
			PlaylistID: "any-playlist-id",
			Tracks: []entities.Song{
				{ID: "any-song-id-1", Title: "any-song-title-1", Album: "any-song-album-1"},
				{ID: "any-song-id-2", Title: "any-song-title-2", Album: "any-song-album-2"},
				{ID: "any-song-id-3", Title: "any-song-title-3", Album: "any-song-album-3"},
			},
		}

		expected := entities.AddTracksToPlaylistClientInput{
			PlaylistID: input.PlaylistID,
			Tracks:     input.Tracks,
		}

		s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return(nil)
		s.ClientMock.On("AddTracksToPlaylist", mock.Anything, expected).Return(nil)

		err := s.UseCase.Execute(context.Background(), input)

		s.NoError(err)
	})

	s.Run("should return error when adding tracks to playlist", func() {
		defer s.cleanMocks()

		input := entities.AddTracksToPlaylistInput{
			PlaylistID: "any-playlist-id",
			Tracks: []entities.Song{
				{ID: "any-song-id-1", Title: "any-song-title-1", Album: "any-song-album-1"},
				{ID: "any-song-id-2", Title: "any-song-title-2", Album: "any-song-album-2"},
				{ID: "any-song-id-3", Title: "any-song-title-3", Album: "any-song-album-3"},
			},
		}

		s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return(nil)
		s.ClientMock.
			On("AddTracksToPlaylist", mock.Anything, mock.Anything).
			Return(errors.New("any-error"))

		err := s.UseCase.Execute(context.Background(), input)

		s.Error(err)
		s.ErrorContains(err, "any-error")
	})
}
