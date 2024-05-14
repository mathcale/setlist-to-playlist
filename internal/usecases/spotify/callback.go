package spotify

import (
	"context"
	"fmt"
	"net/http"

	"github.com/zmb3/spotify/v2"

	spotifyclient "github.com/mathcale/setlist-to-playlist/internal/clients/spotify"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/logger"
	oauth2util "github.com/mathcale/setlist-to-playlist/internal/pkg/oauth2"
)

type SpotifyAuthCallbackUseCaseInterface interface {
	Execute(ctx context.Context, r *http.Request, state string, genCodes oauth2util.GenerateOutput) (*spotify.Client, error)
}

type SpotifyAuthCallbackUseCase struct {
	Client spotifyclient.SpotifyClientInterface
	Logger logger.LoggerInterface
}

func NewSpotifyAuthCallbackUseCase(
	c spotifyclient.SpotifyClientInterface,
	l logger.LoggerInterface,
) SpotifyAuthCallbackUseCaseInterface {
	return &SpotifyAuthCallbackUseCase{
		Client: c,
		Logger: l,
	}
}

func (uc *SpotifyAuthCallbackUseCase) Execute(
	ctx context.Context,
	r *http.Request,
	state string,
	genCodes oauth2util.GenerateOutput,
) (*spotify.Client, error) {
	uc.Logger.Debug("Fetching Spotify token", map[string]interface{}{
		"state":     state,
		"pkceCodes": genCodes,
	})

	token, err := uc.Client.GetToken(ctx, r, state, genCodes)
	if err != nil {
		return nil, err
	}

	uc.Logger.Debug("Spotify token fetched", map[string]interface{}{
		"token": token,
	})

	if st := r.FormValue("state"); st != state {
		return nil, fmt.Errorf("state mismatch: %s != %s", st, state)
	}

	return uc.Client.NewAPIClient(ctx, token), nil
}
