package di

import (
	"github.com/mathcale/setlist-to-playlist/config"
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
}

func NewDependencyInjector(c *config.Config) *DependencyInjector {
	return &DependencyInjector{
		Config: c,
	}
}

func (di *DependencyInjector) Inject() (*Dependencies, error) {
	pkceGen := pkce.NewPKCECodeGenerator()

	return &Dependencies{
		PKCECodeGenerator: pkceGen,
	}, nil
}
