package persistence

import (
	"encoding/json"

	"github.com/mathcale/setlist-to-playlist/internal/entities/applemusic"
	"github.com/mathcale/setlist-to-playlist/internal/infra/persistence/strategies"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/logger"
)

type AppleMusicAuthPersistenceInterface interface {
	Read() (*applemusic.AuthData, error)
	Write(data applemusic.AuthData) error
}

type AppleMusicAuthPersistence struct {
	Strategy strategies.PersistenceStrategyInterface
	Logger   logger.LoggerInterface
}

func NewAppleMusicAuthPersistence(
	strategy strategies.PersistenceStrategyInterface,
	logger logger.LoggerInterface,
) AppleMusicAuthPersistenceInterface {
	return &AppleMusicAuthPersistence{Strategy: strategy, Logger: logger}
}

func (p *AppleMusicAuthPersistence) Read() (*applemusic.AuthData, error) {
	data, err := p.Strategy.Read()
	if err != nil {
		return nil, err
	}

	var authData applemusic.AuthData
	if err := json.Unmarshal(data, &authData); err != nil {
		return nil, err
	}

	return &authData, nil
}

func (p *AppleMusicAuthPersistence) Write(data applemusic.AuthData) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return p.Strategy.Write(dataBytes)
}
