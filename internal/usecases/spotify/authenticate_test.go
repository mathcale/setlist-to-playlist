package spotify

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	entities "github.com/mathcale/setlist-to-playlist/internal/entities/spotify"
	oauth2util "github.com/mathcale/setlist-to-playlist/internal/pkg/oauth2"
	"github.com/mathcale/setlist-to-playlist/internal/tests/mocks"
)

type SpotifyUserAuthenticationUseCaseTestSuite struct {
	suite.Suite
	GatewayMock *mocks.SpotifyUserAuthenticationUseCaseGatewayMock

	UseCase SpotifyUserAuthenticationUseCaseInterface
}

func (s *SpotifyUserAuthenticationUseCaseTestSuite) SetupTest() {
	s.GatewayMock = new(mocks.SpotifyUserAuthenticationUseCaseGatewayMock)

	s.UseCase = NewSpotifyUserAuthenticationUseCase(s.GatewayMock)
}

func (s *SpotifyUserAuthenticationUseCaseTestSuite) cleanMocks() {
	s.GatewayMock.ExpectedCalls = nil
	s.GatewayMock.Calls = nil
}

func TestSpotifyUserAuthenticationUseCase(t *testing.T) {
	suite.Run(t, new(SpotifyUserAuthenticationUseCaseTestSuite))
}

func (s *SpotifyUserAuthenticationUseCaseTestSuite) TestExecute() {
	s.Run("should authenticate user", func() {
		defer s.cleanMocks()

		state := "any-state"
		pkceCodes := oauth2util.GenerateOutput{
			CodeVerifier:  "any-code-verifier",
			CodeChallenge: "any-code-challenge",
		}

		s.GatewayMock.On("ValidatePersistedToken").Return(nil, errors.New("any-error"))
		s.GatewayMock.On("AuthenticateUser", state, pkceCodes).Return(nil)

		err := s.UseCase.Execute(nil, pkceCodes, state)

		s.NoError(err)
	})

	s.Run("should refresh token", func() {
		defer s.cleanMocks()

		ctx := context.Background()

		authData := entities.NewSpotifyUserAuthData(
			"any-access-token",
			"9999-12-31T23:59:59Z",
			"any-refresh-token",
			"any-token-type",
		)

		pkceCodes := oauth2util.GenerateOutput{
			CodeVerifier:  "any-code-verifier",
			CodeChallenge: "any-code-challenge",
		}

		s.GatewayMock.On("ValidatePersistedToken").Return(&authData, nil)
		s.GatewayMock.On("RefreshToken", ctx, &authData).Return(nil)

		err := s.UseCase.Execute(ctx, pkceCodes, "any-state")

		s.NoError(err)
	})

	s.Run("should return error when refreshing token", func() {
		defer s.cleanMocks()

		ctx := context.Background()

		authData := entities.NewSpotifyUserAuthData(
			"any-access-token",
			"9999-12-31T23:59:59Z",
			"any-refresh-token",
			"any-token-type",
		)

		pkceCodes := oauth2util.GenerateOutput{
			CodeVerifier:  "any-code-verifier",
			CodeChallenge: "any-code-challenge",
		}

		s.GatewayMock.On("ValidatePersistedToken").Return(&authData, nil)
		s.GatewayMock.On("RefreshToken", ctx, &authData).Return(errors.New("any-error"))

		err := s.UseCase.Execute(ctx, pkceCodes, "any-state")

		s.Error(err)
		s.ErrorContains(err, "any-error")
	})

	s.Run("should return error when authenticating user", func() {
		defer s.cleanMocks()

		state := "any-state"
		pkceCodes := oauth2util.GenerateOutput{
			CodeVerifier:  "any-code-verifier",
			CodeChallenge: "any-code-challenge",
		}

		s.GatewayMock.On("ValidatePersistedToken").Return(nil, errors.New("any-error"))
		s.GatewayMock.On("AuthenticateUser", state, pkceCodes).Return(errors.New("any-error"))

		err := s.UseCase.Execute(nil, pkceCodes, state)

		s.Error(err)
		s.ErrorContains(err, "any-error")
	})
}
