package gateways

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/browser"

	client "github.com/mathcale/setlist-to-playlist/internal/clients/spotify"
	entities "github.com/mathcale/setlist-to-playlist/internal/entities/spotify"
	"github.com/mathcale/setlist-to-playlist/internal/infra/persistence"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/logger"
	oauth2util "github.com/mathcale/setlist-to-playlist/internal/pkg/oauth2"
)

type SpotifyUserAuthenticationUseCaseGatewayInterface interface {
	ValidatePersistedToken() (*entities.SpotifyUserAuthData, error)
	AuthenticateUser(state string, pkceCodes oauth2util.GenerateOutput) error
	RefreshToken(ctx context.Context, authData *entities.SpotifyUserAuthData) error
}

type SpotifyUserAuthenticationUseCaseGateway struct {
	Client                     client.SpotifyClientInterface
	Persistence                persistence.SpotifyAuthPersistenceInterface
	Logger                     logger.LoggerInterface
	AuthenticatedClientChannel chan client.AuthenticatedClient
}

func NewSpotifyUserAuthenticationUseCaseGateway(
	c client.SpotifyClientInterface,
	p persistence.SpotifyAuthPersistenceInterface,
	l logger.LoggerInterface,
	ch chan client.AuthenticatedClient,
) SpotifyUserAuthenticationUseCaseGatewayInterface {
	return &SpotifyUserAuthenticationUseCaseGateway{
		Client:                     c,
		Persistence:                p,
		Logger:                     l,
		AuthenticatedClientChannel: ch,
	}
}

func (gw *SpotifyUserAuthenticationUseCaseGateway) ValidatePersistedToken() (*entities.SpotifyUserAuthData, error) {
	authData, err := gw.Persistence.Read()
	if err != nil {
		return nil, err
	}

	return authData, authData.Validate()
}

func (gw *SpotifyUserAuthenticationUseCaseGateway) AuthenticateUser(
	state string,
	pkceCodes oauth2util.GenerateOutput,
) error {
	gw.Logger.Debug("Starting authentication process.", nil)

	authURL := gw.Client.GetAuthURL(state, pkceCodes)

	gw.Logger.Info(
		fmt.Sprintf("Opening browser for Spotify authentication.\nIf nothing happens, please visit the following URL: %s", authURL),
		nil,
	)

	if err := browser.OpenURL(authURL); err != nil {
		gw.Logger.Warn("Failed to open browser, use the link above to proceed", nil)
	}

	gw.Client.SetAuthenticatedClient(gw.AuthenticatedClientChannel)

	if err := gw.persistToken(); err != nil {
		return err
	}

	return nil
}

func (gw *SpotifyUserAuthenticationUseCaseGateway) RefreshToken(
	ctx context.Context,
	authData *entities.SpotifyUserAuthData,
) error {
	token, err := authData.ToOauth2Token()
	if err != nil {
		return err
	}

	newToken, err := gw.Client.RefreshToken(ctx, token)
	if err != nil {
		return err
	}

	cl := gw.Client.NewAPIClient(ctx, newToken)

	gw.Client.SetAuthenticatedClientFromInstance(client.AuthenticatedClient{
		Client: *cl,
	})

	if err := gw.persistToken(); err != nil {
		return err
	}

	return nil
}

func (gw *SpotifyUserAuthenticationUseCaseGateway) persistToken() error {
	token, err := gw.Client.CurrentSession()
	if err != nil {
		return err
	}

	gw.Logger.Debug("Persisting Spotify session token locally", map[string]interface{}{
		"token": token,
	})

	authData := entities.NewSpotifyUserAuthData(
		token.AccessToken,
		token.Expiry.Format(time.RFC3339),
		token.RefreshToken,
		token.TokenType,
	)

	if err := gw.Persistence.Write(authData); err != nil {
		return err
	}

	gw.Logger.Debug("Spotify session token successfully saved", nil)

	return nil
}
