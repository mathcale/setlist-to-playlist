package spotify

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	entities "github.com/mathcale/setlist-to-playlist/internal/entities/spotify"
	"github.com/mathcale/setlist-to-playlist/internal/tests/mocks"
)

type CreatePlaylistUseCaseTestSuite struct {
	suite.Suite
	ClientMock *mocks.SpotifyClientMock
	LoggerMock *mocks.LoggerMock

	UseCase CreatePlaylistUseCaseInterface
}

func (s *CreatePlaylistUseCaseTestSuite) SetupTest() {
	s.ClientMock = new(mocks.SpotifyClientMock)
	s.LoggerMock = new(mocks.LoggerMock)

	s.UseCase = NewCreatePlaylistUseCase(
		s.ClientMock,
		s.LoggerMock,
	)
}

func (s *CreatePlaylistUseCaseTestSuite) cleanMocks() {
	s.ClientMock.ExpectedCalls = nil
	s.ClientMock.Calls = nil
	s.LoggerMock.ExpectedCalls = nil
	s.LoggerMock.Calls = nil
}

func TestCreatePlaylistUseCase(t *testing.T) {
	suite.Run(t, new(CreatePlaylistUseCaseTestSuite))
}

func (s *CreatePlaylistUseCaseTestSuite) TestExecute() {
	s.Run("should create playlist", func() {
		defer s.cleanMocks()

		description := "any-description"

		input := entities.CreatePlaylistInput{
			Title:       "any-title",
			Description: &description,
		}

		expected := &entities.CreatePlaylistOutput{
			ID:  "any-playlist-id",
			URL: "any-playlist-url",
		}

		s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return()
		s.ClientMock.
			On("CreatePlaylist", mock.Anything, input.Title, *input.Description).
			Return(expected, nil)

		out, err := s.UseCase.Execute(context.Background(), input)

		s.NoError(err)
		s.Equal(expected, out)
	})

	s.Run("should return error when creating playlist", func() {
		defer s.cleanMocks()

		description := "any-description"

		input := entities.CreatePlaylistInput{
			Title:       "any-title",
			Description: &description,
		}

		s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return()
		s.ClientMock.
			On("CreatePlaylist", mock.Anything, input.Title, *input.Description).
			Return(nil, errors.New("any-error"))

		out, err := s.UseCase.Execute(context.Background(), input)

		s.Error(err)
		s.ErrorContains(err, "any-error")
		s.Nil(out)
	})
}
