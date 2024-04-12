package simpletoken

import "github.com/Heatdog/Avito/pkg/token"

const (
	userToken  = "user_token"
	adminToken = "admin_token"
)

type Provider struct{}

func NewSimpleTokenProvider() token.Provider {
	return &Provider{}
}

func (provider Provider) VerifyToken(token string) bool {
	return token == userToken || token == adminToken
}

func (provider Provider) VerifyOnAdmin(token string) bool {
	return token == adminToken
}
