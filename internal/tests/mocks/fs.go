package mocks

import (
	"io/fs"
	"os"

	"github.com/stretchr/testify/mock"
)

type FileSystemDriverMock struct {
	mock.Mock
}

func (m *FileSystemDriverMock) Read(path string) ([]byte, error) {
	args := m.Called(path)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *FileSystemDriverMock) Write(path string, data []byte, perm fs.FileMode) error {
	args := m.Called(path, data, perm)
	return args.Error(0)
}

func (m *FileSystemDriverMock) CreateFile(path string) (*os.File, error) {
	args := m.Called(path)
	return args.Get(0).(*os.File), args.Error(1)
}

func (m *FileSystemDriverMock) CreateDir(path string, perm fs.FileMode) error {
	args := m.Called(path, perm)
	return args.Error(0)
}

func (m *FileSystemDriverMock) Exists(path string) bool {
	args := m.Called(path)
	return args.Bool(0)
}
