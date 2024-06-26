package persistence

import (
	"encoding/json"

	"github.com/mathcale/setlist-to-playlist/internal/entities/spotify"
	"github.com/mathcale/setlist-to-playlist/internal/infra/persistence/strategies"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/logger"
)

type SpotifyAuthPersistenceInterface interface {
	Read() (*spotify.SpotifyUserAuthData, error)
	Write(data spotify.SpotifyUserAuthData) error
}

type SpotifyAuthPersistence struct {
	Strategy strategies.PersistenceStrategyInterface
	Logger   logger.LoggerInterface
}

func NewSpotifyAuthPersistence(
	strategy strategies.PersistenceStrategyInterface,
	logger logger.LoggerInterface,
) SpotifyAuthPersistenceInterface {
	return &SpotifyAuthPersistence{
		Strategy: strategy,
		Logger:   logger,
	}
}

func (p *SpotifyAuthPersistence) Read() (*spotify.SpotifyUserAuthData, error) {
	data, err := p.Strategy.Read()
	if err != nil {
		return nil, err
	}

	var authData spotify.SpotifyUserAuthData
	if err := json.Unmarshal(data, &authData); err != nil {
		return nil, err
	}

	return &authData, nil
}

func (p *SpotifyAuthPersistence) Write(data spotify.SpotifyUserAuthData) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return p.Strategy.Write(dataBytes)
}
