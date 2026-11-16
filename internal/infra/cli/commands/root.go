package commands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mathcale/setlist-to-playlist/internal/entities/music"
	"github.com/mathcale/setlist-to-playlist/internal/infra/cli/commands/gateways"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/logger"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/prompter"
)

type RootCmdInterface interface {
	Build() *cobra.Command
}

type RootCmd struct {
	Logger   logger.LoggerInterface
	Gateway  gateways.RootCmdGatewayInterface
	Prompter prompter.Prompter
	// IsTerminal reports whether stdin is interactive; it is a seam so tests can fake it.
	IsTerminal IsTerminalFunc
}

func NewRootCmd(
	l logger.LoggerInterface,
	gw gateways.RootCmdGatewayInterface,
	p prompter.Prompter,
	isTerminal IsTerminalFunc,
) RootCmdInterface {
	return &RootCmd{
		Logger:     l,
		Gateway:    gw,
		Prompter:   p,
		IsTerminal: isTerminal,
	}
}

func (s *RootCmd) Build() *cobra.Command {
	cmd := &cobra.Command{
		Short: "Creates a playlist based on a Setlist.fm entry",
		RunE:  s.run,
	}

	cmd.Flags().String("url", "", "setlist.fm set URL to create a playlist from")
	cmd.Flags().StringP(
		"provider", "p", string(music.DefaultProvider),
		"music provider to create the playlist on (spotify, apple-music)",
	)
	cmd.MarkFlagRequired("url")

	return cmd
}

func (rc *RootCmd) run(cmd *cobra.Command, args []string) error {
	setlistfmURL, _ := cmd.Flags().GetString("url")

	provider, err := rc.resolveProvider(cmd)
	if err != nil {
		return err
	}

	name := provider.DisplayName()

	if err := rc.Gateway.CheckProvider(provider); err != nil {
		if errors.Is(err, music.ErrProviderUnavailable) {
			rc.Logger.Info(fmt.Sprintf("%s support is not available yet, try another provider", name), nil)
			return nil
		}

		return err
	}

	rc.Logger.Info("Fetching setlist...", nil)

	set, err := rc.Gateway.GetTracksFromSetlist(setlistfmURL)
	if err != nil {
		rc.Logger.Error("Failed to get tracks from setlist", err, nil)
		return err
	}

	rc.Logger.Info(fmt.Sprintf("Authenticating on %s...", name), nil)

	if err := rc.Gateway.Authenticate(cmd.Context(), provider); err != nil {
		rc.Logger.Error(fmt.Sprintf("Failed to authenticate on %s", name), err, nil)
		return err
	}

	rc.Logger.Info(fmt.Sprintf("Fetching songs on %s...", name), nil)

	songs, err := rc.Gateway.FetchSongs(cmd.Context(), provider, set.Songs(), set.ArtistName())
	if err != nil {
		rc.Logger.Error(fmt.Sprintf("Failed to fetch songs from %s", name), err, nil)
		return err
	}

	if len(songs.Skipped) > 0 {
		rc.Logger.Warn("Some songs were not found and were skipped", map[string]interface{}{
			"skipped": strings.Join(songs.Skipped, ", "),
		})
	}

	rc.Logger.Info(fmt.Sprintf("Creating playlist on %s...", name), nil)

	playlistURL, err := rc.Gateway.CreatePlaylist(cmd.Context(), provider, set.Title(), songs.Songs)
	if err != nil {
		rc.Logger.Error(fmt.Sprintf("Failed to create playlist on %s", name), err, nil)
		return err
	}

	rc.Logger.Info(fmt.Sprintf("Playlist created successfully, check it out: %s", *playlistURL), nil)
	return nil
}

// resolveProvider reads --provider. When the flag is absent it shows a select
// menu if stdin is a terminal, and otherwise defaults to Spotify so non-interactive
// runs never hang.
func (rc *RootCmd) resolveProvider(cmd *cobra.Command) (music.ProviderID, error) {
	if cmd.Flags().Changed("provider") {
		raw, _ := cmd.Flags().GetString("provider")
		return music.ParseProviderID(raw)
	}

	if rc.IsTerminal == nil || !rc.IsTerminal() {
		return music.DefaultProvider, nil
	}

	ids := music.ProviderIDs()
	options := make([]prompter.Option, 0, len(ids))
	for _, id := range ids {
		options = append(options, prompter.Option{Label: id.DisplayName(), Value: string(id)})
	}

	choice, err := rc.Prompter.Select("Which music provider do you want to use?", options)
	if err != nil {
		return "", err
	}

	return music.ParseProviderID(choice)
}
