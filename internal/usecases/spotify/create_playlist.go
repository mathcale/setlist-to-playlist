package spotify

import (
	"context"

	client "github.com/mathcale/setlist-to-playlist/internal/clients/spotify"
	entities "github.com/mathcale/setlist-to-playlist/internal/entities/spotify"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/logger"
)

type CreatePlaylistUseCaseInterface interface {
	Execute(ctx context.Context, input entities.CreatePlaylistInput) (*entities.CreatePlaylistOutput, error)
}

type CreatePlaylistUseCase struct {
	Client client.SpotifyClientInterface
	Logger logger.LoggerInterface
}

func NewCreatePlaylistUseCase(
	c client.SpotifyClientInterface,
	l logger.LoggerInterface,
) CreatePlaylistUseCaseInterface {
	return &CreatePlaylistUseCase{
		Client: c,
		Logger: l,
	}
}

func (uc *CreatePlaylistUseCase) Execute(
	ctx context.Context,
	input entities.CreatePlaylistInput,
) (*entities.CreatePlaylistOutput, error) {
	uc.Logger.Debug("Creating playlist on Spotify", map[string]interface{}{
		"title":       input.Title,
		"description": input.Description,
	})

	output, err := uc.Client.CreatePlaylist(ctx, input.Title, input.GetDescription())
	if err != nil {
		return nil, err
	}

	uc.Logger.Debug("Playlist created on Spotify", map[string]interface{}{
		"playlist_id": output.ID,
	})

	return output, nil
}
