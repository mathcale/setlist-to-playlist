package di

import (
	"time"

	"github.com/dchest/uniuri"

	"github.com/mathcale/setlist-to-playlist/config"
	applemusic_client "github.com/mathcale/setlist-to-playlist/internal/clients/applemusic"
	"github.com/mathcale/setlist-to-playlist/internal/clients/setlistfm"
	spotify_client "github.com/mathcale/setlist-to-playlist/internal/clients/spotify"
	"github.com/mathcale/setlist-to-playlist/internal/entities/music"
	"github.com/mathcale/setlist-to-playlist/internal/infra/cli"
	"github.com/mathcale/setlist-to-playlist/internal/infra/cli/commands"
	rootcmd_gw "github.com/mathcale/setlist-to-playlist/internal/infra/cli/commands/gateways"
	"github.com/mathcale/setlist-to-playlist/internal/infra/persistence"
	"github.com/mathcale/setlist-to-playlist/internal/infra/persistence/drivers"
	"github.com/mathcale/setlist-to-playlist/internal/infra/persistence/strategies/plaintext"
	"github.com/mathcale/setlist-to-playlist/internal/infra/web"
	spotify_handlers "github.com/mathcale/setlist-to-playlist/internal/infra/web/handlers/spotify"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/httpclient"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/logger"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/oauth2"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/prompter"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/responsehandler"
	applemusic_ucs "github.com/mathcale/setlist-to-playlist/internal/usecases/applemusic"
	setlistfm_ucs "github.com/mathcale/setlist-to-playlist/internal/usecases/setlistfm"
	spotify_ucs "github.com/mathcale/setlist-to-playlist/internal/usecases/spotify"
	spotify_uc_gw "github.com/mathcale/setlist-to-playlist/internal/usecases/spotify/gateways"
)

type DependencyInjectorInterface interface {
	Inject() (*Dependencies, error)
}

type DependencyInjector struct {
	Config      *config.Config
	ConfigPaths config.ConfigPaths
	Prompter    prompter.Prompter
}

type Dependencies struct {
	CLI cli.CLIInterface
}

func NewDependencyInjector(
	c *config.Config,
	configPaths config.ConfigPaths,
	p prompter.Prompter,
) *DependencyInjector {
	return &DependencyInjector{
		Config:      c,
		ConfigPaths: configPaths,
		Prompter:    p,
	}
}

func (di *DependencyInjector) Inject() (*Dependencies, error) {
	ch := make(chan spotify_client.AuthenticatedClient)
	state := uniuri.New()

	fsDriver := drivers.NewFileSystemDriver()

	pkceGen := oauth2.NewPKCECodeGenerator()
	genCodes, err := pkceGen.Generate()
	if err != nil {
		return nil, err
	}

	l := logger.NewLogger(di.Config.LogLevel)
	responseHandler := responsehandler.NewWebResponseHandler()

	setlistFMHttpClient := httpclient.NewHttpClient(
		di.Config.SetlistFM.BaseURL,
		time.Duration(di.Config.SetlistFM.Timeout)*time.Millisecond,
	)

	setlistFMClient := setlistfm.NewSetlistFMClient(
		setlistFMHttpClient,
		di.Config.SetlistFM.APIKey,
	)

	// Spotify credentials are asked for on first Spotify use (by the
	// authenticator); the client reads them lazily after that.
	spotifyCredentials := config.NewSpotifyCredentialStore(di.Prompter)

	spotifyClient := spotify_client.NewSpotifyClient(
		l,
		di.Config.Spotify.RedirectURL,
		spotifyCredentials,
	)

	plainTextPersistence := plaintext.NewPlainTextPersistenceStrategy(
		fsDriver,
		l,
		di.ConfigPaths.SpotifyAuthFile,
	)

	spotifyAuthPersistence := persistence.NewSpotifyAuthPersistence(plainTextPersistence, l)

	spotifyUserAuthenticationUseCaseGateway := spotify_uc_gw.
		NewSpotifyUserAuthenticationUseCaseGateway(
			spotifyClient,
			spotifyAuthPersistence,
			l,
			ch,
		)

	spotifyCallbackUseCase := spotify_ucs.NewSpotifyAuthCallbackUseCase(spotifyClient, l)
	getSetlistByIDUseCase := setlistfm_ucs.NewGetSetlistByIDUseCase(setlistFMClient)
	spotifyUserAuthenticationUseCase := spotify_ucs.NewSpotifyUserAuthenticationUseCase(
		spotifyUserAuthenticationUseCaseGateway,
	)

	spotifyCallbackHandler := spotify_handlers.NewSpotifyAuthCallbackWebHandler(
		l,
		spotifyCallbackUseCase,
		responseHandler,
		*genCodes,
		state,
		ch,
	)

	webRouter := web.NewWebRouter(spotifyCallbackHandler)
	webServer := web.NewWebServer(di.Config.WebServerPort, l, webRouter.Build())

	spotifyAuthenticator := spotify_ucs.NewSpotifyAuthenticator(
		spotifyCredentials,
		webServer,
		spotifyUserAuthenticationUseCase,
		*genCodes,
		state,
	)

	// Apple Music needs no web server and never touches the Spotify
	// credential store: it asks for its own tokens on first use.
	appleMusicHTTPClient := httpclient.NewHttpClient(
		applemusic_client.BaseURL,
		time.Duration(di.Config.SetlistFM.Timeout)*time.Millisecond,
	)
	appleMusicClient := applemusic_client.NewClient(appleMusicHTTPClient)
	appleMusicAuthPersistence := persistence.NewAppleMusicAuthPersistence(
		plaintext.NewPlainTextPersistenceStrategy(fsDriver, l, di.ConfigPaths.AppleMusicAuthFile),
		l,
	)
	appleMusicAuthenticator := applemusic_ucs.NewAuthenticator(
		appleMusicAuthPersistence, appleMusicClient, di.Prompter, l,
	)

	// Registering a provider is one line.
	providers := music.NewProviderRegistry().
		Register(music.ProviderSpotify, spotifyClient, spotifyAuthenticator).
		Register(music.ProviderAppleMusic, appleMusicClient, appleMusicAuthenticator)

	rootCmdGw := rootcmd_gw.NewRootCmdGateway(l, providers, getSetlistByIDUseCase)

	rootCmd := commands.NewRootCmd(l, rootCmdGw)
	cli := cli.NewCLI(rootCmd.Build())

	return &Dependencies{
		CLI: cli,
	}, nil
}
