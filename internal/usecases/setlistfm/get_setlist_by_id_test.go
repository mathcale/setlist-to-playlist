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

		in := entity.NewGetSetlistByIDInput(
			"https://www.setlist.fm/setlist/blink182/2024/autodromo-de-interlagos-sao-paulo-brazil-53aa1325.html",
		)
		setlistID, _ := in.SetlistID()

		expected := &entity.Set{
			ID: "53aa1325",
		}

		s.ClientMock.On("GetSetlistByID", *setlistID).Return(expected, nil)

		result, err := s.UseCase.Execute(in)

		s.NoError(err)
		s.Equal(expected, result)
	})

	s.Run("Should return an error while fetching setlist", func() {
		defer s.cleanMocks()

		in := entity.NewGetSetlistByIDInput(
			"https://www.setlist.fm/setlist/blink182/2024/autodromo-de-interlagos-sao-paulo-brazil-53aa1325.html",
		)
		setlistID, _ := in.SetlistID()

		s.ClientMock.On("GetSetlistByID", *setlistID).Return(nil, errors.New("any-error"))

		result, err := s.UseCase.Execute(in)

		s.Error(err)
		s.ErrorContains(err, "any-error")
		s.Nil(result)
	})
}
