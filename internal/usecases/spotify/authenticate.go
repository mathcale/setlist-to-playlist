package spotify

import (
	"context"

	oauth2util "github.com/mathcale/setlist-to-playlist/internal/pkg/oauth2"
	"github.com/mathcale/setlist-to-playlist/internal/usecases/spotify/gateways"
)

type SpotifyUserAuthenticationUseCaseInterface interface {
	Execute(ctx context.Context, pkceCodes oauth2util.GenerateOutput, state string) error
}

type SpotifyUserAuthenticationUseCase struct {
	Gateway gateways.SpotifyUserAuthenticationUseCaseGatewayInterface
}

func NewSpotifyUserAuthenticationUseCase(
	gw gateways.SpotifyUserAuthenticationUseCaseGatewayInterface,
) SpotifyUserAuthenticationUseCaseInterface {
	return &SpotifyUserAuthenticationUseCase{
		Gateway: gw,
	}
}

func (uc *SpotifyUserAuthenticationUseCase) Execute(
	ctx context.Context,
	pkceCodes oauth2util.GenerateOutput,
	state string,
) error {
	if authData, err := uc.Gateway.ValidatePersistedToken(); err == nil {
		return uc.Gateway.RefreshToken(ctx, authData)
	}

	return uc.Gateway.AuthenticateUser(state, pkceCodes)
}
