package spotify

import (
	"context"

	client "github.com/mathcale/setlist-to-playlist/internal/clients/spotify"
	entities "github.com/mathcale/setlist-to-playlist/internal/entities/spotify"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/logger"
)

type AddTracksToPlaylistUseCaseInterface interface {
	Execute(ctx context.Context, input entities.AddTracksToPlaylistInput) error
}

type AddTracksToPlaylistUseCase struct {
	Client client.SpotifyClientInterface
	Logger logger.LoggerInterface
}

func NewAddTracksToPlaylistUseCase(
	c client.SpotifyClientInterface,
	l logger.LoggerInterface,
) AddTracksToPlaylistUseCaseInterface {
	return &AddTracksToPlaylistUseCase{
		Client: c,
		Logger: l,
	}
}

func (uc *AddTracksToPlaylistUseCase) Execute(
	ctx context.Context,
	input entities.AddTracksToPlaylistInput,
) error {
	uc.Logger.Debug("Adding tracks to playlist on Spotify", map[string]interface{}{
		"playlist_id": input.PlaylistID,
		"tracks":      input.Tracks,
	})

	err := uc.Client.AddTracksToPlaylist(ctx, entities.AddTracksToPlaylistClientInput{
		PlaylistID: input.PlaylistID,
		Tracks:     input.Tracks,
	})
	if err != nil {
		return err
	}

	uc.Logger.Debug("Tracks added to playlist on Spotify", nil)

	return nil
}
