package gateways

import (
	"context"

	"github.com/mathcale/setlist-to-playlist/internal/entities/music"
	"github.com/mathcale/setlist-to-playlist/internal/entities/setlistfm"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/logger"
	music_ucs "github.com/mathcale/setlist-to-playlist/internal/usecases/music"
	setlistfm_ucs "github.com/mathcale/setlist-to-playlist/internal/usecases/setlistfm"
)

type RootCmdGatewayInterface interface {
	GetTracksFromSetlist(setlistfmURL string) (*setlistfm.Set, error)
	Authenticate(context.Context) error
	FetchSongs(ctx context.Context, songTitles []string, artist string) (*music.FetchSongsOutput, error)
	CreatePlaylist(ctx context.Context, playlistName string, songs []music.Song) (*string, error)
}

type RootCmdGateway struct {
	Logger                logger.LoggerInterface
	Authenticator         music.Authenticator
	GetSetlistByIDUseCase setlistfm_ucs.GetSetlistByIDUseCaseInterface
	FetchSongsUseCase     music_ucs.FetchSongsUseCaseInterface
	CreatePlaylistUseCase music_ucs.CreatePlaylistUseCaseInterface
	AddTracksUseCase      music_ucs.AddTracksUseCaseInterface
}

func NewRootCmdGateway(
	logger logger.LoggerInterface,
	authenticator music.Authenticator,
	getSetlistByIDUseCase setlistfm_ucs.GetSetlistByIDUseCaseInterface,
	fetchSongsUseCase music_ucs.FetchSongsUseCaseInterface,
	createPlaylistUseCase music_ucs.CreatePlaylistUseCaseInterface,
	addTracksUseCase music_ucs.AddTracksUseCaseInterface,
) RootCmdGatewayInterface {
	return &RootCmdGateway{
		Logger:                logger,
		Authenticator:         authenticator,
		GetSetlistByIDUseCase: getSetlistByIDUseCase,
		FetchSongsUseCase:     fetchSongsUseCase,
		CreatePlaylistUseCase: createPlaylistUseCase,
		AddTracksUseCase:      addTracksUseCase,
	}
}

func (gw *RootCmdGateway) GetTracksFromSetlist(setlistfmURL string) (*setlistfm.Set, error) {
	gw.Logger.Debug("Extracting Setlist.fm ID from URL", map[string]interface{}{
		"url": setlistfmURL,
	})

	in := setlistfm.NewGetSetlistByIDInput(setlistfmURL)

	set, err := gw.GetSetlistByIDUseCase.Execute(in)
	if err != nil {
		return nil, err
	}

	gw.Logger.Debug("Setlist loaded", map[string]interface{}{
		"set": set,
	})

	return set, nil
}

func (gw *RootCmdGateway) Authenticate(ctx context.Context) error {
	return gw.Authenticator.Authenticate(ctx)
}

func (gw *RootCmdGateway) FetchSongs(
	ctx context.Context,
	songTitles []string,
	artist string,
) (*music.FetchSongsOutput, error) {
	return gw.FetchSongsUseCase.Execute(ctx, songTitles, artist)
}

func (gw *RootCmdGateway) CreatePlaylist(
	ctx context.Context,
	playlistName string,
	songs []music.Song,
) (*string, error) {
	createPlaylistOut, err := gw.CreatePlaylistUseCase.Execute(
		ctx,
		music.CreatePlaylistInput{
			Title:       playlistName,
			Description: nil,
		},
	)
	if err != nil {
		return nil, err
	}

	gw.Logger.Debug("Adding songs to playlist...", map[string]interface{}{
		"playlistID":  createPlaylistOut.ID,
		"playlistURL": createPlaylistOut.URL,
	})

	if err := gw.AddTracksUseCase.Execute(
		ctx,
		music.AddTracksInput{
			PlaylistID: createPlaylistOut.ID,
			Tracks:     songs,
		},
	); err != nil {
		return nil, err
	}

	return &createPlaylistOut.URL, nil
}
