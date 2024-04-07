package simple_token

import "github.com/Heatdog/Avito/pkg/token"

const (
	userToken  = "user_token"
	adminToken = "admin_token"
)

type SimpleTokenProvider struct{}

func NewSimpleTokenProvider() token.TokenProvider {
	return &SimpleTokenProvider{}
}

func (provider SimpleTokenProvider) VerifyToken(token string) bool {
	return token == userToken || token == adminToken
}

func (provider SimpleTokenProvider) VerifyOnAdmin(token string) bool {
	return token == adminToken
}
