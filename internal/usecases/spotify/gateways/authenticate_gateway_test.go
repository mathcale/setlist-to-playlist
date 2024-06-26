package gateways

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/zmb3/spotify/v2"

	client "github.com/mathcale/setlist-to-playlist/internal/clients/spotify"
	entity "github.com/mathcale/setlist-to-playlist/internal/entities/spotify"
	oauth2util "github.com/mathcale/setlist-to-playlist/internal/pkg/oauth2"
	"github.com/mathcale/setlist-to-playlist/internal/tests/mocks"
)

type SpotifyUserAuthenticationUseCaseGatewayTestSuite struct {
	suite.Suite
	ClientMock                 *mocks.SpotifyClientMock
	PersistenceMock            *mocks.SpotifyAuthPersistenceMock
	LoggerMock                 *mocks.LoggerMock
	AuthenticatedClientChannel chan client.AuthenticatedClient

	Gateway SpotifyUserAuthenticationUseCaseGatewayInterface
}

func (s *SpotifyUserAuthenticationUseCaseGatewayTestSuite) SetupTest() {
	s.ClientMock = new(mocks.SpotifyClientMock)
	s.PersistenceMock = new(mocks.SpotifyAuthPersistenceMock)
	s.LoggerMock = new(mocks.LoggerMock)
	s.AuthenticatedClientChannel = make(chan client.AuthenticatedClient)

	s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return(nil)
	s.LoggerMock.On("Info", mock.Anything, mock.Anything).Return(nil)
	s.LoggerMock.On("Warn", mock.Anything, mock.Anything).Return(nil)
	s.LoggerMock.On("Error", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	s.Gateway = NewSpotifyUserAuthenticationUseCaseGateway(
		s.ClientMock,
		s.PersistenceMock,
		s.LoggerMock,
		s.AuthenticatedClientChannel,
	)
}

func (s *SpotifyUserAuthenticationUseCaseGatewayTestSuite) cleanMocks() {
	s.ClientMock.ExpectedCalls = nil
	s.ClientMock.Calls = nil
	s.PersistenceMock.ExpectedCalls = nil
	s.PersistenceMock.Calls = nil
	s.LoggerMock.ExpectedCalls = nil
	s.LoggerMock.Calls = nil
}

func TestSpotifyUserAuthenticationUseCaseGateway(t *testing.T) {
	suite.Run(t, new(SpotifyUserAuthenticationUseCaseGatewayTestSuite))
}

func (s *SpotifyUserAuthenticationUseCaseGatewayTestSuite) TestValidatePersistedToken() {
	s.Run("Should return valid entity", func() {
		defer s.cleanMocks()

		authData := entity.NewSpotifyUserAuthData(
			"any-access-token",
			"9999-12-31T23:59:59Z",
			"any-refresh-token",
			"any-token-type",
		)

		s.PersistenceMock.On("Read").Return(&authData, nil)

		result, err := s.Gateway.ValidatePersistedToken()

		s.NoError(err)
		s.Equal(&authData, result)
	})

	s.Run("Should return error when persistence read fails", func() {
		defer s.cleanMocks()

		anyError := errors.New("any-error")
		s.PersistenceMock.On("Read").Return(nil, anyError)

		_, err := s.Gateway.ValidatePersistedToken()

		s.ErrorContains(err, anyError.Error())
	})
}

func (s *SpotifyUserAuthenticationUseCaseGatewayTestSuite) TestAuthenticateUser() {
	s.Run("Should authenticate user", func() {
		defer s.cleanMocks()

		state := "any-state"
		pkceCodes := oauth2util.GenerateOutput{
			CodeChallenge: "any-code-challenge",
			CodeVerifier:  "any-code-verifier",
		}
		authData, _ := entity.NewSpotifyUserAuthData(
			"any-access-token",
			"9999-12-31T23:59:59Z",
			"any-refresh-token",
			"any-token-type",
		).ToOauth2Token()

		authURL := "any-auth-url"

		s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return(nil)
		s.LoggerMock.On("Info", mock.Anything, mock.Anything).Return(nil)
		s.LoggerMock.On("Warn", mock.Anything, mock.Anything).Return(nil)
		s.ClientMock.On("GetAuthURL", state, pkceCodes).Return(authURL)
		s.ClientMock.On("SetAuthenticatedClient", mock.Anything).Return(nil)
		s.ClientMock.On("CurrentSession").Return(authData, nil)
		s.PersistenceMock.On("Write", mock.Anything).Return(nil)

		go func() {
			<-s.AuthenticatedClientChannel
		}()

		err := s.Gateway.AuthenticateUser(state, pkceCodes)

		s.NoError(err)

	})

	s.Run("Should return error while persisting token", func() {
		defer s.cleanMocks()

		state := "any-state"
		pkceCodes := oauth2util.GenerateOutput{
			CodeChallenge: "any-code-challenge",
			CodeVerifier:  "any-code-verifier",
		}
		authData, _ := entity.NewSpotifyUserAuthData(
			"any-access-token",
			"9999-12-31T23:59:59Z",
			"any-refresh-token",
			"any-token-type",
		).ToOauth2Token()

		authURL := "any-auth-url"

		s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return(nil)
		s.LoggerMock.On("Info", mock.Anything, mock.Anything).Return(nil)
		s.LoggerMock.On("Warn", mock.Anything, mock.Anything).Return(nil)
		s.ClientMock.On("GetAuthURL", state, pkceCodes).Return(authURL)
		s.ClientMock.On("SetAuthenticatedClient", mock.Anything).Return(nil)
		s.ClientMock.On("CurrentSession").Return(authData, nil)
		s.PersistenceMock.On("Write", mock.Anything).Return(errors.New("any-error"))

		go func() {
			<-s.AuthenticatedClientChannel
		}()

		err := s.Gateway.AuthenticateUser(state, pkceCodes)

		s.Error(err)
	})
}

func (s *SpotifyUserAuthenticationUseCaseGatewayTestSuite) TestRefreshToken() {
	s.Run("Should refresh token", func() {
		defer s.cleanMocks()

		authData := entity.NewSpotifyUserAuthData(
			"any-access-token",
			"9999-12-31T23:59:59Z",
			"any-refresh-token",
			"any-token-type",
		)

		tok, _ := authData.ToOauth2Token()

		s.ClientMock.On("RefreshToken", mock.Anything, mock.Anything).Return(nil, nil)
		s.ClientMock.On("NewAPIClient", mock.Anything, mock.Anything).Return(&spotify.Client{})
		s.ClientMock.On("SetAuthenticatedClientFromInstance", mock.Anything).Return(nil)
		s.ClientMock.On("CurrentSession").Return(tok, nil)
		s.PersistenceMock.On("Write", mock.Anything).Return(nil)

		err := s.Gateway.RefreshToken(context.Background(), &authData)

		s.NoError(err)
	})

	s.Run("Should return error while refreshing token", func() {
		defer s.cleanMocks()

		authData := entity.NewSpotifyUserAuthData(
			"any-access-token",
			"2024-01-31T23:59:59Z",
			"any-refresh-token",
			"any-token-type",
		)

		err := errors.New("any-error")

		s.ClientMock.On("RefreshToken", mock.Anything, mock.Anything).Return(nil, err)

		err = s.Gateway.RefreshToken(context.Background(), &authData)

		s.Error(err)
		s.ErrorContains(err, "any-error")
	})

	s.Run("Should return error while persisting token", func() {
		defer s.cleanMocks()

		authData := entity.NewSpotifyUserAuthData(
			"any-access-token",
			"9999-01-31T23:59:59Z",
			"any-refresh-token",
			"any-token-type",
		)

		tok, _ := authData.ToOauth2Token()

		s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return(nil)
		s.ClientMock.On("RefreshToken", mock.Anything, mock.Anything).Return(tok, nil)
		s.ClientMock.On("NewAPIClient", mock.Anything, mock.Anything).Return(&spotify.Client{})
		s.ClientMock.On("SetAuthenticatedClientFromInstance", mock.Anything).Return(nil)
		s.ClientMock.On("CurrentSession").Return(tok, nil)
		s.PersistenceMock.On("Write", mock.Anything).Return(errors.New("any-write-error"))

		err := s.Gateway.RefreshToken(context.Background(), &authData)

		s.Error(err)
		s.ErrorContains(err, "any-write-error")
	})
}
