package music

import (
	"context"

	"github.com/mathcale/setlist-to-playlist/internal/entities/music"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/logger"
)

type CreatePlaylistUseCaseInterface interface {
	Execute(ctx context.Context, input music.CreatePlaylistInput) (*music.Playlist, error)
}

type CreatePlaylistUseCase struct {
	Provider music.MusicProvider
	Logger   logger.LoggerInterface
}

func NewCreatePlaylistUseCase(p music.MusicProvider, l logger.LoggerInterface) CreatePlaylistUseCaseInterface {
	return &CreatePlaylistUseCase{Provider: p, Logger: l}
}

func (uc *CreatePlaylistUseCase) Execute(
	ctx context.Context,
	input music.CreatePlaylistInput,
) (*music.Playlist, error) {
	uc.Logger.Debug("Creating playlist", map[string]interface{}{
		"title":       input.Title,
		"description": input.Description,
	})

	playlist, err := uc.Provider.CreatePlaylist(ctx, input.Title, input.GetDescription())
	if err != nil {
		return nil, err
	}

	uc.Logger.Debug("Playlist created", map[string]interface{}{"playlist_id": playlist.ID})

	return playlist, nil
}
