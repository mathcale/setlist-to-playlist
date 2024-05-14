package mocks

import (
	"github.com/stretchr/testify/mock"

	"github.com/mathcale/setlist-to-playlist/internal/entities/setlistfm"
)

type SetlistFMExtractIDFromURLUseCaseMock struct {
	mock.Mock
}

type SetlistFMGetSetlistByIDUseCaseMock struct {
	mock.Mock
}

type SetlistFMClientMock struct {
	mock.Mock
}

func (m *SetlistFMExtractIDFromURLUseCaseMock) Execute(url string) (*string, error) {
	args := m.Called(url)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*string), args.Error(1)
}

func (m *SetlistFMGetSetlistByIDUseCaseMock) Execute(id string) (*setlistfm.Set, error) {
	args := m.Called(id)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*setlistfm.Set), args.Error(1)
}

func (m *SetlistFMClientMock) GetSetlistByID(setlistID string) (*setlistfm.Set, error) {
	args := m.Called(setlistID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*setlistfm.Set), args.Error(1)
}
