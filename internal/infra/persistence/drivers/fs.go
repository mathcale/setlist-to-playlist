package drivers

import (
	"io/fs"
	"os"
)

type FileSystemDriverInterface interface {
	Read(path string) ([]byte, error)
	Write(path string, data []byte, perm fs.FileMode) error
	CreateFile(path string) (*os.File, error)
	CreateDir(path string, perm fs.FileMode) error
	Exists(path string) bool
}

type FileSystemDriver struct{}

func NewFileSystemDriver() FileSystemDriverInterface {
	return &FileSystemDriver{}
}

func (d *FileSystemDriver) Read(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (d *FileSystemDriver) Write(path string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(path, data, perm)
}

func (d *FileSystemDriver) CreateFile(path string) (*os.File, error) {
	return os.Create(path)
}

func (d *FileSystemDriver) CreateDir(path string, perm fs.FileMode) error {
	if !d.Exists(path) {
		if err := os.Mkdir(path, perm); err != nil {
			return err
		}
	}

	return nil
}

func (d *FileSystemDriver) Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
