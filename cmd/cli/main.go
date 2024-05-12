package main

import (
	"log"

	"github.com/mathcale/setlist-to-playlist/config"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/di"
)

func main() {
	cfg, err := config.Load(".")
	if err != nil {
		log.Fatalf("There was an error while loading config: %s", err)
	}

	d := di.NewDependencyInjector(cfg)

	_, err = d.Inject()
	if err != nil {
		log.Fatalf("There was an error while injecting dependencies: %s", err)
	}
}
