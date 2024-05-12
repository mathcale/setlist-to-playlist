package di

import (
	"time"

	"github.com/mathcale/setlist-to-playlist/config"
	"github.com/mathcale/setlist-to-playlist/internal/clients/setlistfm"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/httpclient"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/pkce"
)

type DependencyInjectorInterface interface {
	Inject() (*Dependencies, error)
}

type DependencyInjector struct {
	Config *config.Config
}

type Dependencies struct {
	PKCECodeGenerator pkce.PKCECodeGeneratorInterface
	SetlistFMClient   setlistfm.SetlistFMClientInterface
}

func NewDependencyInjector(c *config.Config) *DependencyInjector {
	return &DependencyInjector{
		Config: c,
	}
}

func (di *DependencyInjector) Inject() (*Dependencies, error) {
	pkceGen := pkce.NewPKCECodeGenerator()

	setlistFMHttpClient := httpclient.NewHttpClient(
		di.Config.SetlistFMAPIBaseURL,
		time.Duration(di.Config.SetlistFMAPITimeout)*time.Millisecond,
	)

	setlistFMClient := setlistfm.NewSetlistFMClient(
		setlistFMHttpClient,
		di.Config.SetlistFMAPIKey,
	)

	return &Dependencies{
		PKCECodeGenerator: pkceGen,
		SetlistFMClient:   setlistFMClient,
	}, nil
}
