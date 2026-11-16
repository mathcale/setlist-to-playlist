package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/mathcale/setlist-to-playlist/internal/entities/applemusic"
)

type AppleMusicAuthPersistenceMock struct {
	mock.Mock
}

func (m *AppleMusicAuthPersistenceMock) Read() (*applemusic.AuthData, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*applemusic.AuthData), args.Error(1)
}

func (m *AppleMusicAuthPersistenceMock) Write(data applemusic.AuthData) error {
	return m.Called(data).Error(0)
}

type StorefrontClientMock struct {
	mock.Mock
}

func (m *StorefrontClientMock) SetAuth(auth applemusic.AuthData) {
	m.Called(auth)
}

func (m *StorefrontClientMock) Storefront(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}
