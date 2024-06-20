package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathcale/setlist-to-playlist/internal/infra/cli/commands/gateways"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/logger"
)

type RootCmdInterface interface {
	Build() *cobra.Command
}

type RootCmd struct {
	Logger  logger.LoggerInterface
	Gateway gateways.RootCmdGatewayInterface
}

func NewRootCmd(
	l logger.LoggerInterface,
	gw gateways.RootCmdGatewayInterface,
) RootCmdInterface {
	return &RootCmd{
		Logger:  l,
		Gateway: gw,
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

	rc.Logger.Info("Fetching setlist...", nil)

	set, err := rc.Gateway.GetTracksFromSetlist(setlistfmURL)
	if err != nil {
		rc.Logger.Error("Failed to get tracks from setlist", err, nil)
		return err
	}

	rc.Gateway.StartWebServer()

	if err := rc.Gateway.HandleSpotifyAuthentication(cmd.Context()); err != nil {
		rc.Logger.Error("Failed to authenticate on Spotify", err, nil)
		return err
	}

	rc.Logger.Info("Fetching songs on Spotify...", nil)

	songs, err := rc.Gateway.FetchSongsOnSpotify(cmd.Context(), set.Songs(), set.ArtistName())
	if err != nil {
		rc.Logger.Error("Failed to fetch songs from Spotify", err, nil)
		return err
	}

	rc.Logger.Info("Creating playlist...", nil)

	playlistURL, err := rc.Gateway.CreatePlaylistOnSpotify(cmd.Context(), set.Title(), songs.Songs)
	if err != nil {
		rc.Logger.Error("Failed to create playlist on Spotify", err, nil)
		return err
	}

	rc.Logger.Info(fmt.Sprintf("Playlist created successfully: %s", *playlistURL), nil)
	return nil
}
