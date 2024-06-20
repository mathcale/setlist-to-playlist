package mocks

import "github.com/stretchr/testify/mock"

type WebServerMock struct {
	mock.Mock
}

func (m *WebServerMock) Start() {
	m.Called()
}
