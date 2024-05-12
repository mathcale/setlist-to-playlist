package setlistfm

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mathcale/setlist-to-playlist/internal/entities/setlistfm"
	"github.com/mathcale/setlist-to-playlist/internal/tests/mocks"
)

type SetlistFMClientTestSuite struct {
	suite.Suite
	HttpClientMock *mocks.HttpClientMock
	APIKey         string

	SetlistFMClient SetlistFMClientInterface
}

func (s *SetlistFMClientTestSuite) SetupTest() {
	s.HttpClientMock = new(mocks.HttpClientMock)
	s.APIKey = "any-api-key"
	s.SetlistFMClient = NewSetlistFMClient(s.HttpClientMock, s.APIKey)
}

func (s *SetlistFMClientTestSuite) cleanMock() {
	s.HttpClientMock.ExpectedCalls = nil
	s.HttpClientMock.Calls = nil
}

func TestSetlistFMClient(t *testing.T) {
	suite.Run(t, new(SetlistFMClientTestSuite))
}

func (s *SetlistFMClientTestSuite) TestGetSetlistByID() {
	s.Run("Should return a setlist", func() {
		defer s.cleanMock()

		id := "any-setlist-id"

		s.HttpClientMock.On("Get", "/1.0/setlist/"+id, map[string]interface{}{"x-api-key": s.APIKey}, &setlistfm.Set{}).Return(nil)

		result, err := s.SetlistFMClient.GetSetlistByID(id)

		s.NoError(err)
		s.NotNil(result)
	})

	s.Run("Should return an error when http client fails", func() {
		defer s.cleanMock()

		id := "any-setlist-id"
		mockError := errors.New("any-error")

		s.HttpClientMock.On("Get", "/1.0/setlist/"+id, map[string]interface{}{"x-api-key": s.APIKey}, &setlistfm.Set{}).Return(mockError)

		result, err := s.SetlistFMClient.GetSetlistByID(id)

		s.Error(err)
		s.Nil(result)
	})
}
