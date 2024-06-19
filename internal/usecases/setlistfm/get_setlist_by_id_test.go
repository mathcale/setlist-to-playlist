package setlistfm

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	entity "github.com/mathcale/setlist-to-playlist/internal/entities/setlistfm"
	"github.com/mathcale/setlist-to-playlist/internal/tests/mocks"
)

type GetSetlistByIDUseCaseTestSuite struct {
	suite.Suite
	ClientMock *mocks.SetlistFMClientMock

	UseCase GetSetlistByIDUseCaseInterface
}

func (s *GetSetlistByIDUseCaseTestSuite) SetupTest() {
	s.ClientMock = new(mocks.SetlistFMClientMock)

	s.UseCase = NewGetSetlistByIDUseCase(s.ClientMock)
}

func (s *GetSetlistByIDUseCaseTestSuite) cleanMocks() {
	s.ClientMock.ExpectedCalls = nil
	s.ClientMock.Calls = nil
}

func TestGetSetlistByIDUseCase(t *testing.T) {
	suite.Run(t, new(GetSetlistByIDUseCaseTestSuite))
}

func (s *GetSetlistByIDUseCaseTestSuite) TestExecute() {
	s.Run("Should return a setlist", func() {
		defer s.cleanMocks()

		id := "any-setlist-id"

		expected := &entity.Set{
			ID: "any-setlist-id",
		}

		s.ClientMock.On("GetSetlistByID", id).Return(expected, nil)

		result, err := s.UseCase.Execute(id)

		s.NoError(err)
		s.Equal(expected, result)
	})

	s.Run("Should return an error while fetching setlist", func() {
		defer s.cleanMocks()

		id := "any-setlist-id"

		s.ClientMock.On("GetSetlistByID", id).Return(nil, errors.New("any-error"))

		result, err := s.UseCase.Execute(id)

		s.Error(err)
		s.ErrorContains(err, "any-error")
		s.Nil(result)
	})
}
