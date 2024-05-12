package mocks

import "github.com/stretchr/testify/mock"

type HttpClientMock struct {
	mock.Mock
}

func (m *HttpClientMock) Get(url string, headers map[string]interface{}, response interface{}) error {
	args := m.Called(url, headers, response)
	return args.Error(0)
}
