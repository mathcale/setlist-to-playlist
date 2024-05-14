package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zmb3/spotify/v2"

	spotifyclient "github.com/mathcale/setlist-to-playlist/internal/clients/spotify"
	"github.com/mathcale/setlist-to-playlist/internal/infra/web"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/logger"
	oauth2util "github.com/mathcale/setlist-to-playlist/internal/pkg/oauth2"
	"github.com/mathcale/setlist-to-playlist/internal/usecases/setlistfm"
)

type RootCmdInterface interface {
	Build() *cobra.Command
}

type RootCmd struct {
	WebServer                        web.WebServerInterface
	Logger                           logger.LoggerInterface
	SpotifyClient                    spotifyclient.SpotifyClientInterface
	ExtractSetlistFMIDFromURLUseCase setlistfm.ExtractIDFromURLUseCaseInterface
	GetSetlistByIDUseCase            setlistfm.GetSetlistByIDUseCaseInterface
	GeneratedPKCECodes               oauth2util.GenerateOutput
	State                            string
	SpotifyClientChannel             chan *spotify.Client
}

func NewRootCmd(
	webServer web.WebServerInterface,
	l logger.LoggerInterface,
	spotifyClient spotifyclient.SpotifyClientInterface,
	extractSetlistFMIDFromURLUseCase setlistfm.ExtractIDFromURLUseCaseInterface,
	getSetlistByIDUseCase setlistfm.GetSetlistByIDUseCaseInterface,
	genCodes oauth2util.GenerateOutput,
	state string,
	ch chan *spotify.Client,
) RootCmdInterface {
	return &RootCmd{
		WebServer:                        webServer,
		Logger:                           l,
		SpotifyClient:                    spotifyClient,
		ExtractSetlistFMIDFromURLUseCase: extractSetlistFMIDFromURLUseCase,
		GetSetlistByIDUseCase:            getSetlistByIDUseCase,
		GeneratedPKCECodes:               genCodes,
		State:                            state,
		SpotifyClientChannel:             ch,
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

	rc.WebServer.Start()

	authURL := rc.SpotifyClient.GetAuthURL(rc.State, rc.GeneratedPKCECodes)
	rc.Logger.Info(fmt.Sprintf("Please visit the following URL to authenticate with Spotify: %s", authURL), nil)

	spotifyClient := <-rc.SpotifyClientChannel

	currUser, err := spotifyClient.CurrentUser(cmd.Context())
	if err != nil {
		rc.Logger.Error("Failed to get current user", err, nil)
		return err
	}

	rc.Logger.Info(fmt.Sprintf("Hello \"%s\"", currUser.Email), nil)

	setlistID, err := rc.ExtractSetlistFMIDFromURLUseCase.Execute(setlistfmURL)
	if err != nil {
		rc.Logger.Error("Failed to extract Setlist.fm ID from URL", err, map[string]interface{}{
			"url": setlistfmURL,
		})

		return err
	}

	rc.Logger.Info(fmt.Sprintf("Loading data from setlist [%s]", *setlistID), nil)

	set, err := rc.GetSetlistByIDUseCase.Execute(*setlistID)
	if err != nil {
		rc.Logger.Error("Failed to get setlist from Setlist.fm", err, map[string]interface{}{
			"setlistID": *setlistID,
		})

		return err
	}

	rc.Logger.Info(fmt.Sprintf("Setlist [%s] loaded", *setlistID), nil)
	rc.Logger.Debug("Setlist data", map[string]interface{}{
		"setlist": set,
	})

	return nil
}
