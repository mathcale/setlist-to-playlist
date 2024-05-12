package pkce

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type PKCECodeGeneratorTestSuite struct {
	suite.Suite
	PKCECodeGenerator PKCECodeGeneratorInterface
}

func (s *PKCECodeGeneratorTestSuite) SetupTest() {
	s.PKCECodeGenerator = NewPKCECodeGenerator()
}

func TestPKCECodeGenerator(t *testing.T) {
	suite.Run(t, new(PKCECodeGeneratorTestSuite))
}

func (s *PKCECodeGeneratorTestSuite) TestGenerate() {
	s.Run("should return code verifier and code challenge", func() {
		output, err := s.PKCECodeGenerator.Generate()

		s.NoError(err)
		s.NotNil(output.CodeVerifier)
		s.NotNil(output.CodeChallenge)
	})
}
