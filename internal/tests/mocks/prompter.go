package mocks

import (
	"github.com/stretchr/testify/mock"

	"github.com/mathcale/setlist-to-playlist/internal/pkg/prompter"
)

type PrompterMock struct {
	mock.Mock
}

func (m *PrompterMock) Text(title string) (string, error) {
	args := m.Called(title)
	return args.String(0), args.Error(1)
}

func (m *PrompterMock) Secret(title string) (string, error) {
	args := m.Called(title)
	return args.String(0), args.Error(1)
}

func (m *PrompterMock) Select(title string, options []prompter.Option) (string, error) {
	args := m.Called(title, options)
	return args.String(0), args.Error(1)
}
