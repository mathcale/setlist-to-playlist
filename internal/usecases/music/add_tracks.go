package music

import (
	"context"

	"github.com/mathcale/setlist-to-playlist/internal/entities/music"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/logger"
)

type AddTracksUseCaseInterface interface {
	Execute(ctx context.Context, input music.AddTracksInput) error
}

type AddTracksUseCase struct {
	Provider music.MusicProvider
	Logger   logger.LoggerInterface
}

func NewAddTracksUseCase(p music.MusicProvider, l logger.LoggerInterface) AddTracksUseCaseInterface {
	return &AddTracksUseCase{Provider: p, Logger: l}
}

func (uc *AddTracksUseCase) Execute(ctx context.Context, input music.AddTracksInput) error {
	uc.Logger.Debug("Adding tracks to playlist", map[string]interface{}{
		"playlist_id": input.PlaylistID,
		"tracks":      input.Tracks,
	})

	ids := make([]string, len(input.Tracks))
	for i, t := range input.Tracks {
		ids[i] = t.ID
	}

	if err := uc.Provider.AddTracks(ctx, input.PlaylistID, ids); err != nil {
		return err
	}

	uc.Logger.Debug("Tracks added to playlist", nil)

	return nil
}
