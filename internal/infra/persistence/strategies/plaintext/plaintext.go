package plaintext

import (
	"github.com/mathcale/setlist-to-playlist/internal/infra/persistence/drivers"
	"github.com/mathcale/setlist-to-playlist/internal/pkg/logger"
)

type PlainTextPersistenceStrategyInterface interface {
	Read() ([]byte, error)
	Write(data []byte) error
}

type PlainTextPersistenceStrategy struct {
	FSDriver drivers.FileSystemDriverInterface
	Logger   logger.LoggerInterface
	Path     string
}

func NewPlainTextPersistenceStrategy(
	d drivers.FileSystemDriverInterface,
	l logger.LoggerInterface,
	path string,
) PlainTextPersistenceStrategyInterface {
	return &PlainTextPersistenceStrategy{
		FSDriver: d,
		Logger:   l,
		Path:     path,
	}
}

func (p *PlainTextPersistenceStrategy) Read() ([]byte, error) {
	p.Logger.Debug("Reading data from file", map[string]interface{}{
		"path": p.Path,
	})

	data, err := p.FSDriver.Read(p.Path)
	if err != nil {
		return nil, err
	}

	p.Logger.Debug("Data read from file", map[string]interface{}{
		"data": string(data),
	})

	return data, nil
}

func (p *PlainTextPersistenceStrategy) Write(data []byte) error {
	p.Logger.Debug("Writing data to file", map[string]interface{}{
		"path": p.Path,
		"data": string(data),
	})

	if err := p.FSDriver.Write(p.Path, data, 0644); err != nil {
		return err
	}

	p.Logger.Debug("Data written to file successfully", nil)

	return nil
}
