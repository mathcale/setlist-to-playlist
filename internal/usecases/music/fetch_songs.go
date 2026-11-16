package music

import (
	"context"

	"github.com/mathcale/setlist-to-playlist/internal/entities/music"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/logger"
)

type FetchSongsUseCaseInterface interface {
	Execute(ctx context.Context, songs []string, artist string) (*music.FetchSongsOutput, error)
}

type FetchSongsUseCase struct {
	Provider music.MusicProvider
	Logger   logger.LoggerInterface
}

func NewFetchSongsUseCase(p music.MusicProvider, l logger.LoggerInterface) FetchSongsUseCaseInterface {
	return &FetchSongsUseCase{Provider: p, Logger: l}
}

func (uc *FetchSongsUseCase) Execute(
	ctx context.Context,
	songs []string,
	artist string,
) (*music.FetchSongsOutput, error) {
	uc.Logger.Debug("Fetching songs", map[string]interface{}{
		"songs":  songs,
		"artist": artist,
	})

	out := &music.FetchSongsOutput{Artist: artist}

	for _, title := range songs {
		candidates, err := uc.Provider.SearchCandidates(ctx, title, artist)
		if err != nil {
			return nil, err
		}

		if len(candidates) == 0 {
			out.Skipped = append(out.Skipped, title)
			continue
		}

		// First candidate wins until ranking is introduced.
		c := candidates[0]
		out.Songs = append(out.Songs, music.Song{ID: c.TrackID, Title: c.Title, Album: c.Album})
	}

	uc.Logger.Debug("Songs fetched", map[string]interface{}{"songs": out})

	return out, nil
}
