package setlistfm

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type GetSetlistByIDInputTestSuite struct {
	suite.Suite

	ValidInput    GetSetlistByIDInput
	WrongURLInput GetSetlistByIDInput
	NoIDInput     GetSetlistByIDInput
	EmptyInput    GetSetlistByIDInput
}

func (s *GetSetlistByIDInputTestSuite) SetupTest() {
	s.ValidInput = NewGetSetlistByIDInput("https://www.setlist.fm/setlist/blink182/2024/autodromo-de-interlagos-sao-paulo-brazil-53aa1325.html")
	s.WrongURLInput = NewGetSetlistByIDInput("https://www.setlist.fm/festival/2024/download-festival-2024-73d44e99.html")
	s.NoIDInput = NewGetSetlistByIDInput("https://www.setlist.fm/festival/2024/download-festival-2024.html")
	s.EmptyInput = NewGetSetlistByIDInput("")
}

func TestGetSetlistByIDInput(t *testing.T) {
	suite.Run(t, new(GetSetlistByIDInputTestSuite))
}

func (s *GetSetlistByIDInputTestSuite) TestSetlistID() {
	s.Run("Should return the setlist ID", func() {
		result, err := s.ValidInput.SetlistID()

		s.NoError(err)
		s.Equal("53aa1325", *result)
	})

	s.Run("Should return an error when URL is empty", func() {
		result, err := s.EmptyInput.SetlistID()

		s.Error(err)
		s.ErrorContains(err, "URL is empty")
		s.Nil(result)
	})

	s.Run("Should return an error when URL is not a setlist.fm set URL", func() {
		result, err := s.WrongURLInput.SetlistID()

		s.Error(err)
		s.ErrorContains(err, "URL is not a valid setlist.fm set")
		s.Nil(result)
	})

	s.Run("Should return an error when URL does not contain an ID", func() {
		result, err := s.NoIDInput.SetlistID()

		s.Error(err)
		s.Nil(result)
	})
}
