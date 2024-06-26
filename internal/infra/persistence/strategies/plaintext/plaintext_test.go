package plaintext

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/setlist-to-playlist/internal/tests/mocks"
)

type PlainTextPersistenceStrategyTestSuite struct {
	suite.Suite
	FSDriverMock                 *mocks.FileSystemDriverMock
	LoggerMock                   *mocks.LoggerMock
	FilePath                     string
	PlainTextPersistenceStrategy PlainTextPersistenceStrategyInterface
}

func (s *PlainTextPersistenceStrategyTestSuite) SetupTest() {
	s.FSDriverMock = new(mocks.FileSystemDriverMock)
	s.LoggerMock = new(mocks.LoggerMock)
	s.FilePath = "/tmp/test.json"

	s.PlainTextPersistenceStrategy = NewPlainTextPersistenceStrategy(
		s.FSDriverMock,
		s.LoggerMock,
		s.FilePath,
	)
}

func (s *PlainTextPersistenceStrategyTestSuite) resetMocks() {
	s.FSDriverMock.ExpectedCalls = nil
	s.FSDriverMock.Calls = nil
	s.LoggerMock.ExpectedCalls = nil
	s.LoggerMock.Calls = nil
}

func TestPlainTextPersistenceStrategy(t *testing.T) {
	suite.Run(t, new(PlainTextPersistenceStrategyTestSuite))
}

func (s *PlainTextPersistenceStrategyTestSuite) TestRead() {
	s.Run("Should read data from file", func() {
		defer s.resetMocks()

		expectedData := []byte(`{"access_token":"test","token_type":"Bearer","refresh_token":"test","expiry":"2021-09-01T00:00:00Z"}`)

		s.FSDriverMock.On("Read", s.FilePath).Return(expectedData, nil)
		s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return()

		data, err := s.PlainTextPersistenceStrategy.Read()

		s.Nil(err)
		s.Equal(expectedData, data)
	})

	s.Run("Should return error when reading data from file fails", func() {
		defer s.resetMocks()

		expectedError := errors.New("any-error")

		s.FSDriverMock.On("Read", s.FilePath).Return([]byte{}, expectedError)
		s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return()

		data, err := s.PlainTextPersistenceStrategy.Read()

		s.Error(err)
		s.Equal(expectedError, err)
		s.Nil(data)
	})
}

func (s *PlainTextPersistenceStrategyTestSuite) TestWrite() {
	s.Run("Should write data to file", func() {
		defer s.resetMocks()

		data := []byte(`{"access_token":"test","token_type":"Bearer","refresh_token":"test","expiry":"2021-09-01T00:00:00Z"}`)

		s.FSDriverMock.On("Write", s.FilePath, mock.Anything, mock.Anything).Return(nil)
		s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return()

		err := s.PlainTextPersistenceStrategy.Write(data)

		s.Nil(err)
	})

	s.Run("Should return error when writing data to file fails", func() {
		defer s.resetMocks()

		expectedError := errors.New("any-error")

		s.FSDriverMock.On("Write", s.FilePath, mock.Anything, mock.Anything).Return(expectedError)
		s.LoggerMock.On("Debug", mock.Anything, mock.Anything).Return()

		err := s.PlainTextPersistenceStrategy.Write([]byte{})

		s.Error(err)
		s.Equal(expectedError, err)
	})
}
