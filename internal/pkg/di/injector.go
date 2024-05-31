package di

import (
	"time"

	"github.com/dchest/uniuri"
	"github.com/zmb3/spotify/v2"

	"github.com/mathcale/setlist-to-playlist/config"
	"github.com/mathcale/setlist-to-playlist/internal/clients/setlistfm"
	spotify_client "github.com/mathcale/setlist-to-playlist/internal/clients/spotify"
	"github.com/mathcale/setlist-to-playlist/internal/gateways"
	"github.com/mathcale/setlist-to-playlist/internal/infra/cli"
	"github.com/mathcale/setlist-to-playlist/internal/infra/cli/commands"
	"github.com/mathcale/setlist-to-playlist/internal/infra/web"
	spotify_handlers "github.com/mathcale/setlist-to-playlist/internal/infra/web/handlers/spotify"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/httpclient"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/logger"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/oauth2"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/responsehandler"
	setlistfm_ucs "github.com/mathcale/setlist-to-playlist/internal/usecases/setlistfm"
	spotify_ucs "github.com/mathcale/setlist-to-playlist/internal/usecases/spotify"
)

type DependencyInjectorInterface interface {
	Inject() (*Dependencies, error)
}

type DependencyInjector struct {
	Config *config.Config
}

type Dependencies struct {
	CLI cli.CLIInterface
}

func NewDependencyInjector(c *config.Config) *DependencyInjector {
	return &DependencyInjector{
		Config: c,
	}
}

func (di *DependencyInjector) Inject() (*Dependencies, error) {
	ch := make(chan *spotify.Client)
	state := uniuri.New()

	pkceGen := oauth2.NewPKCECodeGenerator()
	genCodes, err := pkceGen.Generate()
	if err != nil {
		return nil, err
	}

	l := logger.NewLogger(di.Config.LogLevel)
	responseHandler := responsehandler.NewWebResponseHandler()

	setlistFMHttpClient := httpclient.NewHttpClient(
		di.Config.SetlistFMAPIBaseURL,
		time.Duration(di.Config.SetlistFMAPITimeout)*time.Millisecond,
	)

	setlistFMClient := setlistfm.NewSetlistFMClient(
		setlistFMHttpClient,
		di.Config.SetlistFMAPIKey,
	)

	spotifyClient := spotify_client.NewSpotifyClient(
		l,
		di.Config.SpotifyRedirectURL,
		di.Config.SpotifyClientID,
		di.Config.SpotifyClientSecret,
	)

	spotifyCallbackUseCase := spotify_ucs.NewSpotifyAuthCallbackUseCase(spotifyClient, l)
	extractSetlistFMIDFromURLUseCase := setlistfm_ucs.NewExtractIDFromURLUseCase()
	getSetlistByIDUseCase := setlistfm_ucs.NewGetSetlistByIDUseCase(setlistFMClient)
	fetchSongsOnSpotifyUseCase := spotify_ucs.NewFetchSongsOnSpotifyUseCase(spotifyClient, l)
	createPlaylistOnSpotifyUseCase := spotify_ucs.NewCreatePlaylistUseCase(spotifyClient, l)
	addTracksToSpotifyPlaylistUseCase := spotify_ucs.NewAddTracksToPlaylistUseCase(spotifyClient, l)

	spotifyCallbackHandler := spotify_handlers.NewSpotifyAuthCallbackWebHandler(
		spotifyCallbackUseCase,
		responseHandler,
		*genCodes,
		state,
		ch,
	)

	webRouter := web.NewWebRouter(spotifyCallbackHandler)
	webServer := web.NewWebServer(di.Config.WebServerPort, l, webRouter.Build())

	rootCmdGw := gateways.NewRootCmdGateway(
		l,
		webServer,
		spotifyClient,
		extractSetlistFMIDFromURLUseCase,
		getSetlistByIDUseCase,
		fetchSongsOnSpotifyUseCase,
		createPlaylistOnSpotifyUseCase,
		addTracksToSpotifyPlaylistUseCase,
		*genCodes,
		state,
		ch,
	)

	rootCmd := commands.NewRootCmd(l, rootCmdGw)
	cli := cli.NewCLI(rootCmd.Build())

	return &Dependencies{
		CLI: cli,
	}, nil
}
