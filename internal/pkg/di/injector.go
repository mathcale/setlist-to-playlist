package di

import (
	"github.com/mathcale/setlist-to-playlist/config"
)

type DependencyInjectorInterface interface {
	Inject() (*Dependencies, error)
}

type DependencyInjector struct {
	Config *config.Config
}

type Dependencies struct {
}

func NewDependencyInjector(c *config.Config) *DependencyInjector {
	return &DependencyInjector{
		Config: c,
	}
}

func (di *DependencyInjector) Inject() (*Dependencies, error) {
	return &Dependencies{}, nil
}
