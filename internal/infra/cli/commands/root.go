package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zmb3/spotify/v2"

	spotifyclient "github.com/mathcale/setlist-to-playlist/internal/clients/spotify"
	spotify_entities "github.com/mathcale/setlist-to-playlist/internal/entities/spotify"
	"github.com/mathcale/setlist-to-playlist/internal/infra/web"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/logger"
	oauth2util "github.com/mathcale/setlist-to-playlist/internal/pkg/oauth2"
	"github.com/mathcale/setlist-to-playlist/internal/usecases/setlistfm"
	spotify_ucs "github.com/mathcale/setlist-to-playlist/internal/usecases/spotify"
)

type RootCmdInterface interface {
	Build() *cobra.Command
}

type RootCmd struct {
	WebServer                         web.WebServerInterface
	Logger                            logger.LoggerInterface
	SpotifyClient                     spotifyclient.SpotifyClientInterface
	ExtractSetlistFMIDFromURLUseCase  setlistfm.ExtractIDFromURLUseCaseInterface
	GetSetlistByIDUseCase             setlistfm.GetSetlistByIDUseCaseInterface
	FetchSongsOnSpotifyUseCase        spotify_ucs.FetchSongsOnSpotifyUseCaseInterface
	CreatePlaylistOnSpotifyUseCase    spotify_ucs.CreatePlaylistUseCaseInterface
	AddTracksToSpotifyPlaylistUseCase spotify_ucs.AddTracksToPlaylistUseCaseInterface
	GeneratedPKCECodes                oauth2util.GenerateOutput
	State                             string
	SpotifyClientChannel              chan *spotify.Client
}

func NewRootCmd(
	webServer web.WebServerInterface,
	l logger.LoggerInterface,
	spotifyClient spotifyclient.SpotifyClientInterface,
	extractSetlistFMIDFromURLUseCase setlistfm.ExtractIDFromURLUseCaseInterface,
	getSetlistByIDUseCase setlistfm.GetSetlistByIDUseCaseInterface,
	fetchSongsOnSpotifyUseCase spotify_ucs.FetchSongsOnSpotifyUseCaseInterface,
	createPlaylistOnSpotifyUseCase spotify_ucs.CreatePlaylistUseCaseInterface,
	addTracksToSpotifyPlaylistUseCase spotify_ucs.AddTracksToPlaylistUseCaseInterface,
	genCodes oauth2util.GenerateOutput,
	state string,
	ch chan *spotify.Client,
) RootCmdInterface {
	return &RootCmd{
		WebServer:                         webServer,
		Logger:                            l,
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

func (s *RootCmd) Build() *cobra.Command {
	cmd := &cobra.Command{
		Short: "Creates a playlist based on a Setlist.fm entry",
		RunE:  s.run,
	}

	cmd.Flags().String("url", "", "setlist.fm set URL to create a playlist from")
	cmd.MarkFlagRequired("url")

	return cmd
}

func (rc *RootCmd) run(cmd *cobra.Command, args []string) error {
	setlistfmURL, _ := cmd.Flags().GetString("url")

	setlistID, err := rc.ExtractSetlistFMIDFromURLUseCase.Execute(setlistfmURL)
	if err != nil {
		rc.Logger.Error("Failed to extract Setlist.fm ID from URL", err, map[string]interface{}{
			"url": setlistfmURL,
		})

		return err
	}

	rc.Logger.Info(fmt.Sprintf("Loading data from setlist [%s]...", *setlistID), nil)

	set, err := rc.GetSetlistByIDUseCase.Execute(*setlistID)
	if err != nil {
		rc.Logger.Error("Failed to get setlist from Setlist.fm", err, map[string]interface{}{
			"setlistID": *setlistID,
		})

		return err
	}

	rc.Logger.Info(fmt.Sprintf("Setlist loaded: [%s]", set.Title()), nil)

	rc.WebServer.Start()

	authURL := rc.SpotifyClient.GetAuthURL(rc.State, rc.GeneratedPKCECodes)
	rc.Logger.Info(fmt.Sprintf("Please visit the following URL to authenticate with Spotify: %s", authURL), nil)

	rc.SpotifyClient.SetAuthenticatedClient(rc.SpotifyClientChannel)

	close(rc.SpotifyClientChannel)

	user, err := rc.SpotifyClient.CurrentUser(cmd.Context())
	if err != nil {
		rc.Logger.Error("Failed to get current user", err, nil)
		return err
	}

	token, err := rc.SpotifyClient.CurrentSession()
	if err != nil {
		rc.Logger.Error("Failed to get Spotify token", err, nil)
		return err
	}

	rc.Logger.Info(
		fmt.Sprintf(
			"Logged in as \"%s\". Session valid until %s",
			user.Email,
			token.Expiry.Local().Format("2006-01-02 15:04:05"),
		),
		nil,
	)

	rc.Logger.Info("Fetching songs...", nil)

	songs, err := rc.FetchSongsOnSpotifyUseCase.Execute(cmd.Context(), set.Songs(), set.ArtistName())
	if err != nil {
		rc.Logger.Error("Failed to fetch songs from Spotify", err, nil)
		return err
	}

	rc.Logger.Info("Creating playlist...", nil)

	createPlaylistOut, err := rc.CreatePlaylistOnSpotifyUseCase.Execute(
		cmd.Context(),
		spotify_entities.CreatePlaylistInput{
			Title:       set.Title(),
			Description: nil,
		},
	)
	if err != nil {
		rc.Logger.Error("Failed to create playlist", err, nil)
		return err
	}

	rc.Logger.Info("Adding songs to playlist...", nil)

	if err := rc.AddTracksToSpotifyPlaylistUseCase.Execute(
		cmd.Context(),
		spotify_entities.AddTracksToPlaylistInput{
			PlaylistID: createPlaylistOut.ID,
			Tracks:     songs.Songs,
		},
	); err != nil {
		rc.Logger.Error("Failed to add songs to playlist", err, nil)
		return err
	}

	rc.Logger.Info(fmt.Sprintf("Playlist created successfully: %s", createPlaylistOut.URL), nil)

	return nil
}
