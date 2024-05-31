package gateways

import (
	"context"
	"fmt"

	"github.com/zmb3/spotify/v2"

	spotifyclient "github.com/mathcale/setlist-to-playlist/internal/clients/spotify"
	"github.com/mathcale/setlist-to-playlist/internal/entities/setlistfm"
	spotify_entities "github.com/mathcale/setlist-to-playlist/internal/entities/spotify"
	"github.com/mathcale/setlist-to-playlist/internal/infra/web"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/logger"
	oauth2util "github.com/mathcale/setlist-to-playlist/internal/pkg/oauth2"
	setlistfm_ucs "github.com/mathcale/setlist-to-playlist/internal/usecases/setlistfm"
	spotify_ucs "github.com/mathcale/setlist-to-playlist/internal/usecases/spotify"
)

type RootCmdGatewayInterface interface {
	GetTracksFromSetlist(setlistfmURL string) (*setlistfm.Set, error)
	StartWebServer()
	HandleSpotifyAuthentication(context.Context) error
	FetchSongsOnSpotify(ctx context.Context, songTitles []string, artist string) (*spotify_entities.FindAllSongsOutput, error)
	CreatePlaylistOnSpotify(ctx context.Context, playlistName string, songs []spotify_entities.Song) (*string, error)
}

type RootCmdGateway struct {
	Logger                            logger.LoggerInterface
	WebServer                         web.WebServerInterface
	SpotifyClient                     spotifyclient.SpotifyClientInterface
	ExtractSetlistFMIDFromURLUseCase  setlistfm_ucs.ExtractIDFromURLUseCaseInterface
	GetSetlistByIDUseCase             setlistfm_ucs.GetSetlistByIDUseCaseInterface
	FetchSongsOnSpotifyUseCase        spotify_ucs.FetchSongsOnSpotifyUseCaseInterface
	CreatePlaylistOnSpotifyUseCase    spotify_ucs.CreatePlaylistUseCaseInterface
	AddTracksToSpotifyPlaylistUseCase spotify_ucs.AddTracksToPlaylistUseCaseInterface
	GeneratedPKCECodes                oauth2util.GenerateOutput
	State                             string
	SpotifyClientChannel              chan *spotify.Client
}

func NewRootCmdGateway(
	logger logger.LoggerInterface,
	webServer web.WebServerInterface,
	spotifyClient spotifyclient.SpotifyClientInterface,
	extractSetlistFMIDFromURLUseCase setlistfm_ucs.ExtractIDFromURLUseCaseInterface,
	getSetlistByIDUseCase setlistfm_ucs.GetSetlistByIDUseCaseInterface,
	fetchSongsOnSpotifyUseCase spotify_ucs.FetchSongsOnSpotifyUseCaseInterface,
	createPlaylistOnSpotifyUseCase spotify_ucs.CreatePlaylistUseCaseInterface,
	addTracksToSpotifyPlaylistUseCase spotify_ucs.AddTracksToPlaylistUseCaseInterface,
	genCodes oauth2util.GenerateOutput,
	state string,
	ch chan *spotify.Client,
) RootCmdGatewayInterface {
	return &RootCmdGateway{
		Logger:                            logger,
		WebServer:                         webServer,
		SpotifyClient:                     spotifyClient,
		ExtractSetlistFMIDFromURLUseCase:  extractSetlistFMIDFromURLUseCase,
		GetSetlistByIDUseCase:             getSetlistByIDUseCase,
		FetchSongsOnSpotifyUseCase:        fetchSongsOnSpotifyUseCase,
		CreatePlaylistOnSpotifyUseCase:    createPlaylistOnSpotifyUseCase,
		AddTracksToSpotifyPlaylistUseCase: addTracksToSpotifyPlaylistUseCase,
		GeneratedPKCECodes:                genCodes,
		State:                             state,
		SpotifyClientChannel:              ch,
	}
}

func (gw *RootCmdGateway) GetTracksFromSetlist(setlistfmURL string) (*setlistfm.Set, error) {
	gw.Logger.Debug("Extracting Setlist.fm ID from URL", map[string]interface{}{
		"url": setlistfmURL,
	})

	setlistID, err := gw.ExtractSetlistFMIDFromURLUseCase.Execute(setlistfmURL)
	if err != nil {
		return nil, err
	}

	gw.Logger.Debug("Loading data from setlist", map[string]interface{}{
		"setlistID": *setlistID,
	})

	set, err := gw.GetSetlistByIDUseCase.Execute(*setlistID)
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
	authURL := gw.SpotifyClient.GetAuthURL(gw.State, gw.GeneratedPKCECodes)
	gw.Logger.Info(
		fmt.Sprintf("Please visit the following URL to authenticate with Spotify: %s", authURL),
		nil,
	)

	gw.SpotifyClient.SetAuthenticatedClient(gw.SpotifyClientChannel)
	close(gw.SpotifyClientChannel)

	if err := gw.getCurrentSpotifySession(ctx); err != nil {
		return err
	}

	return nil
}

func (gw *RootCmdGateway) FetchSongsOnSpotify(
	ctx context.Context,
	songTitles []string,
	artist string,
) (*spotify_entities.FindAllSongsOutput, error) {
	songs, err := gw.FetchSongsOnSpotifyUseCase.Execute(ctx, songTitles, artist)
	if err != nil {
		return nil, err
	}

	return songs, nil
}

func (gw *RootCmdGateway) CreatePlaylistOnSpotify(
	ctx context.Context,
	playlistName string,
	songs []spotify_entities.Song,
) (*string, error) {
	createPlaylistOut, err := gw.CreatePlaylistOnSpotifyUseCase.Execute(
		ctx,
		spotify_entities.CreatePlaylistInput{
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

	if err := gw.AddTracksToSpotifyPlaylistUseCase.Execute(
		ctx,
		spotify_entities.AddTracksToPlaylistInput{
			PlaylistID: createPlaylistOut.ID,
			Tracks:     songs,
		},
	); err != nil {
		return nil, err
	}

	return &createPlaylistOut.URL, nil
}

func (gw *RootCmdGateway) getCurrentSpotifySession(ctx context.Context) error {
	user, err := gw.SpotifyClient.CurrentUser(ctx)
	if err != nil {
		return err
	}

	token, err := gw.SpotifyClient.CurrentSession()
	if err != nil {
		return err
	}

	gw.Logger.Info(
		fmt.Sprintf(
			"Logged in as \"%s\". Session valid until %s",
			user.Email,
			token.Expiry.Local().Format("2006-01-02 15:04:05"),
		),
		nil,
	)

	return nil
}
