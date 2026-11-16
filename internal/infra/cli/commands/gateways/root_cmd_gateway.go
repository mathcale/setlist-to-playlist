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
	// CheckProvider returns music.ErrProviderUnavailable (wrapped) when the
	// provider has nothing registered. It never authenticates.
	CheckProvider(id music.ProviderID) error
	Authenticate(ctx context.Context, id music.ProviderID) error
	FetchSongs(ctx context.Context, id music.ProviderID, songTitles []string, artist string) (*music.FetchSongsOutput, error)
	CreatePlaylist(ctx context.Context, id music.ProviderID, playlistName string, songs []music.Song) (*string, error)
}

// RootCmdGateway looks up the chosen provider in the registry and builds the
// provider-agnostic music use cases around it per call. The use cases are
// stateless wrappers, so constructing them on demand keeps them decoupled
// from any single provider.
type RootCmdGateway struct {
	Logger                logger.LoggerInterface
	Registry              *music.ProviderRegistry
	GetSetlistByIDUseCase setlistfm_ucs.GetSetlistByIDUseCaseInterface
}

func NewRootCmdGateway(
	logger logger.LoggerInterface,
	registry *music.ProviderRegistry,
	getSetlistByIDUseCase setlistfm_ucs.GetSetlistByIDUseCaseInterface,
) RootCmdGatewayInterface {
	return &RootCmdGateway{
		Logger:                logger,
		Registry:              registry,
		GetSetlistByIDUseCase: getSetlistByIDUseCase,
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

func (gw *RootCmdGateway) CheckProvider(id music.ProviderID) error {
	_, err := gw.Registry.Lookup(id)
	return err
}

func (gw *RootCmdGateway) Authenticate(ctx context.Context, id music.ProviderID) error {
	entry, err := gw.Registry.Lookup(id)
	if err != nil {
		return err
	}

	return music_ucs.NewRecovery(entry.Authenticator, gw.Logger, id).Authenticate(ctx)
}

// recovering wraps the provider so each call recovers once from expired
// credentials (re-authenticate, then retry just that call).
func (gw *RootCmdGateway) recovering(entry music.ProviderEntry, id music.ProviderID) music.MusicProvider {
	return music_ucs.NewRecoveringProvider(entry.Provider, music_ucs.NewRecovery(entry.Authenticator, gw.Logger, id))
}

func (gw *RootCmdGateway) FetchSongs(
	ctx context.Context,
	id music.ProviderID,
	songTitles []string,
	artist string,
) (*music.FetchSongsOutput, error) {
	entry, err := gw.Registry.Lookup(id)
	if err != nil {
		return nil, err
	}

	return music_ucs.NewFetchSongsUseCase(gw.recovering(entry, id), gw.Logger).Execute(ctx, songTitles, artist)
}

func (gw *RootCmdGateway) CreatePlaylist(
	ctx context.Context,
	id music.ProviderID,
	playlistName string,
	songs []music.Song,
) (*string, error) {
	entry, err := gw.Registry.Lookup(id)
	if err != nil {
		return nil, err
	}

	createPlaylistOut, err := music_ucs.NewCreatePlaylistUseCase(gw.recovering(entry, id), gw.Logger).Execute(
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

	if err := music_ucs.NewAddTracksUseCase(gw.recovering(entry, id), gw.Logger).Execute(
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
