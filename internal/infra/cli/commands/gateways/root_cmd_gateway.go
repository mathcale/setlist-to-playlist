package gateways

import (
	"context"

	spotifyclient "github.com/mathcale/setlist-to-playlist/internal/clients/spotify"
	"github.com/mathcale/setlist-to-playlist/internal/entities/music"
	"github.com/mathcale/setlist-to-playlist/internal/entities/setlistfm"
	"github.com/mathcale/setlist-to-playlist/internal/infra/web"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/logger"
	oauth2util "github.com/mathcale/setlist-to-playlist/internal/pkg/oauth2"
	music_ucs "github.com/mathcale/setlist-to-playlist/internal/usecases/music"
	setlistfm_ucs "github.com/mathcale/setlist-to-playlist/internal/usecases/setlistfm"
	spotify_ucs "github.com/mathcale/setlist-to-playlist/internal/usecases/spotify"
)

type RootCmdGatewayInterface interface {
	GetTracksFromSetlist(setlistfmURL string) (*setlistfm.Set, error)
	StartWebServer()
	HandleSpotifyAuthentication(context.Context) error
	FetchSongs(ctx context.Context, songTitles []string, artist string) (*music.FetchSongsOutput, error)
	CreatePlaylist(ctx context.Context, playlistName string, songs []music.Song) (*string, error)
}

type RootCmdGateway struct {
	Logger                           logger.LoggerInterface
	WebServer                        web.WebServerInterface
	SpotifyClient                    spotifyclient.SpotifyClientInterface
	GetSetlistByIDUseCase            setlistfm_ucs.GetSetlistByIDUseCaseInterface
	FetchSongsUseCase                music_ucs.FetchSongsUseCaseInterface
	CreatePlaylistUseCase            music_ucs.CreatePlaylistUseCaseInterface
	AddTracksUseCase                 music_ucs.AddTracksUseCaseInterface
	SpotifyUserAuthenticationUseCase spotify_ucs.SpotifyUserAuthenticationUseCaseInterface
	GeneratedPKCECodes               oauth2util.GenerateOutput
	State                            string
	SpotifyClientChannel             chan spotifyclient.AuthenticatedClient
}

func NewRootCmdGateway(
	logger logger.LoggerInterface,
	webServer web.WebServerInterface,
	spotifyClient spotifyclient.SpotifyClientInterface,
	getSetlistByIDUseCase setlistfm_ucs.GetSetlistByIDUseCaseInterface,
	fetchSongsUseCase music_ucs.FetchSongsUseCaseInterface,
	createPlaylistUseCase music_ucs.CreatePlaylistUseCaseInterface,
	addTracksUseCase music_ucs.AddTracksUseCaseInterface,
	spotifyUserAuthenticationUseCase spotify_ucs.SpotifyUserAuthenticationUseCaseInterface,
	genCodes oauth2util.GenerateOutput,
	state string,
	ch chan spotifyclient.AuthenticatedClient,
) RootCmdGatewayInterface {
	return &RootCmdGateway{
		Logger:                           logger,
		WebServer:                        webServer,
		SpotifyClient:                    spotifyClient,
		GetSetlistByIDUseCase:            getSetlistByIDUseCase,
		FetchSongsUseCase:                fetchSongsUseCase,
		CreatePlaylistUseCase:            createPlaylistUseCase,
		AddTracksUseCase:                 addTracksUseCase,
		SpotifyUserAuthenticationUseCase: spotifyUserAuthenticationUseCase,
		GeneratedPKCECodes:               genCodes,
		State:                            state,
		SpotifyClientChannel:             ch,
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

func (gw *RootCmdGateway) StartWebServer() {
	gw.WebServer.Start()
}

func (gw *RootCmdGateway) HandleSpotifyAuthentication(ctx context.Context) error {
	return gw.SpotifyUserAuthenticationUseCase.Execute(ctx, gw.GeneratedPKCECodes, gw.State)
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
