package spotify

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/zmb3/spotify/v2"

	entities "github.com/mathcale/setlist-to-playlist/internal/entities/spotify"
	oauth2util "github.com/mathcale/setlist-to-playlist/internal/pkg/oauth2"
	"github.com/mathcale/setlist-to-playlist/internal/tests/mocks"
)

type SpotifyAuthCallbackUseCaseTestSuite struct {
	suite.Suite
	ClientMock *mocks.SpotifyClientMock
	LoggerMock *mocks.LoggerMock

	UseCase SpotifyAuthCallbackUseCaseInterface
}

func (s *SpotifyAuthCallbackUseCaseTestSuite) SetupTest() {
	s.ClientMock = new(mocks.SpotifyClientMock)
	s.LoggerMock = new(mocks.LoggerMock)

	s.UseCase = NewSpotifyAuthCallbackUseCase(
		s.ClientMock,
		s.LoggerMock,
	)
}

func (s *SpotifyAuthCallbackUseCaseTestSuite) cleanMocks() {
	s.ClientMock.ExpectedCalls = nil
	s.ClientMock.Calls = nil
	s.LoggerMock.ExpectedCalls = nil
	s.LoggerMock.Calls = nil
}

func TestSpotifyAuthCallbackUseCase(t *testing.T) {
	suite.Run(t, new(SpotifyAuthCallbackUseCaseTestSuite))
}

func (s *SpotifyAuthCallbackUseCaseTestSuite) TestExecute() {
	s.Run("should fetch Spotify token from request", func() {
		defer s.cleanMocks()

		state := "any-state"
		genCodes := oauth2util.GenerateOutput{
			CodeVerifier:  "any-code-verifier",
			CodeChallenge: "any-code-challenge",
		}

		tok, _ := entities.NewSpotifyUserAuthData(
			"any-access-token",
			"9999-12-31T23:59:59Z",
			"any-refresh-token",
			"any-token-type",
		).ToOauth2Token()

		r := &http.Request{
			Form: map[string][]string{
				"state": {state},
			},
		}

		client := new(spotify.Client)

		s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return()
		s.ClientMock.On("GetToken", mock.Anything, mock.Anything, state, genCodes).Return(tok, nil)
		s.ClientMock.On("NewAPIClient", mock.Anything, tok).Return(client)

		_, err := s.UseCase.Execute(context.Background(), r, state, genCodes)

		s.NotNil(client)
		s.NoError(err)
	})

	s.Run("should return error when state mismatch", func() {
		defer s.cleanMocks()

		state := "any-state"
		genCodes := oauth2util.GenerateOutput{
			CodeVerifier:  "any-code-verifier",
			CodeChallenge: "any-code-challenge",
		}

		tok, _ := entities.NewSpotifyUserAuthData(
			"any-access-token",
			"9999-12-31T23:59:59Z",
			"any-refresh-token",
			"any-token-type",
		).ToOauth2Token()

		r := &http.Request{
			Form: map[string][]string{
				"state": {"another-state"},
			},
		}

		s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return()
		s.ClientMock.On("GetToken", mock.Anything, mock.Anything, state, genCodes).Return(tok, nil)

		_, err := s.UseCase.Execute(context.Background(), r, state, genCodes)

		s.Error(err)
	})

	s.Run("should return error when fetching Spotify token", func() {
		defer s.cleanMocks()

		state := "any-state"
		genCodes := oauth2util.GenerateOutput{
			CodeVerifier:  "any-code-verifier",
			CodeChallenge: "any-code-challenge",
		}

		r := &http.Request{
			Form: map[string][]string{
				"state": {state},
			},
		}

		s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return()
		s.ClientMock.On("GetToken", mock.Anything, mock.Anything, state, genCodes).Return(nil, errors.New("any-error"))

		_, err := s.UseCase.Execute(context.Background(), r, state, genCodes)

		s.Error(err)
		s.ErrorContains(err, "any-error")
	})
}
