package pkce

import (
	go_oauth_pkce_code_verifier "github.com/nirasan/go-oauth-pkce-code-verifier"
)

type GenerateOutput struct {
	CodeVerifier  string
	CodeChallenge string
}

type PKCECodeGeneratorInterface interface {
	Generate() (*GenerateOutput, error)
}

type PKCECodeGenerator struct{}

func NewPKCECodeGenerator() *PKCECodeGenerator {
	return &PKCECodeGenerator{}
}

func (g *PKCECodeGenerator) Generate() (*GenerateOutput, error) {
	v, err := go_oauth_pkce_code_verifier.CreateCodeVerifier()
	if err != nil {
		return nil, err
	}

	return &GenerateOutput{
		CodeVerifier:  v.String(),
		CodeChallenge: v.CodeChallengeS256(),
	}, nil
}
