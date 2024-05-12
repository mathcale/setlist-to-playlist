package main

import (
	"github.com/mathcale/setlist-to-playlist/config"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/di"
)

func main() {
	cfg, err := config.Load(".")
	if err != nil {
		panic(err)
	}

	d := di.NewDependencyInjector(cfg)

	_, err = d.Inject()
	if err != nil {
		panic(err)
	}
}
