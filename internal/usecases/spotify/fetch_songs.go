package spotify

import (
	"context"

	client "github.com/mathcale/setlist-to-playlist/internal/clients/spotify"
	entities "github.com/mathcale/setlist-to-playlist/internal/entities/spotify"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/logger"
)

type FetchSongsOnSpotifyUseCaseInterface interface {
	Execute(ctx context.Context, songs []string, artist string) (*entities.FindAllSongsOutput, error)
}

type FetchSongsOnSpotifyUseCase struct {
	Client client.SpotifyClientInterface
	Logger logger.LoggerInterface
}

func NewFetchSongsOnSpotifyUseCase(c client.SpotifyClientInterface, l logger.LoggerInterface) FetchSongsOnSpotifyUseCaseInterface {
	return &FetchSongsOnSpotifyUseCase{
		Client: c,
		Logger: l,
	}
}

func (uc *FetchSongsOnSpotifyUseCase) Execute(
	ctx context.Context,
	songs []string,
	artist string,
) (*entities.FindAllSongsOutput, error) {
	uc.Logger.Debug("Fetching songs on Spotify", map[string]interface{}{
		"songs":  songs,
		"artist": artist,
	})

	output, err := uc.Client.FindAllSongsByName(ctx, songs, artist)
	if err != nil {
		return nil, err
	}

	uc.Logger.Debug("Original songs", map[string]interface{}{
		"songs": songs,
	})

	uc.Logger.Debug("Songs fetched from Spotify", map[string]interface{}{
		"output.songs": output.Songs,
	})

	output.SortSongs(songs)

	return output, nil
}
