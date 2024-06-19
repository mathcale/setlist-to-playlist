package spotify

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	dto "github.com/mathcale/setlist-to-playlist/internal/entities/spotify"
	"github.com/mathcale/setlist-to-playlist/internal/tests/mocks"
)

type FetchSongsOnSpotifyUseCaseTestSuite struct {
	suite.Suite
	ClientMock *mocks.SpotifyClientMock
	LoggerMock *mocks.LoggerMock

	UseCase FetchSongsOnSpotifyUseCaseInterface
}

func (s *FetchSongsOnSpotifyUseCaseTestSuite) SetupTest() {
	s.ClientMock = new(mocks.SpotifyClientMock)
	s.LoggerMock = new(mocks.LoggerMock)

	s.UseCase = NewFetchSongsOnSpotifyUseCase(
		s.ClientMock,
		s.LoggerMock,
	)
}

func (s *FetchSongsOnSpotifyUseCaseTestSuite) cleanMocks() {
	s.ClientMock.ExpectedCalls = nil
	s.ClientMock.Calls = nil
	s.LoggerMock.ExpectedCalls = nil
	s.LoggerMock.Calls = nil
}

func TestFetchSongsOnSpotifyUseCase(t *testing.T) {
	suite.Run(t, new(FetchSongsOnSpotifyUseCaseTestSuite))
}

func (s *FetchSongsOnSpotifyUseCaseTestSuite) TestExecute() {
	s.Run("should fetch songs from Spotify", func() {
		defer s.cleanMocks()

		songs := []string{"any-song-title-1", "any-song-title-2", "any-song-title-3"}
		artist := "any-artist"

		expected := &dto.FindAllSongsOutput{
			Songs: []dto.Song{
				{ID: "any-song-id-1", Title: "any-song-title-1", Album: "any-song-album-1"},
				{ID: "any-song-id-2", Title: "any-song-title-2", Album: "any-song-album-2"},
				{ID: "any-song-id-3", Title: "any-song-title-3", Album: "any-song-album-3"},
			},
		}

		s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return()
		s.ClientMock.On("FindAllSongsByName", mock.Anything, songs, artist).Return(expected, nil)

		out, err := s.UseCase.Execute(context.Background(), songs, artist)

		s.NoError(err)
		s.Equal(expected, out)
	})

	s.Run("should return error when fetching songs from Spotify", func() {
		defer s.cleanMocks()

		songs := []string{"any-song-title-1", "any-song-title-2", "any-song-title-3"}
		artist := "any-artist"

		s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return()
		s.ClientMock.
			On("FindAllSongsByName", mock.Anything, songs, artist).
			Return(nil, errors.New("any-error"))

		out, err := s.UseCase.Execute(context.Background(), songs, artist)

		s.Error(err)
		s.ErrorContains(err, "any-error")
		s.Nil(out)
	})
}
