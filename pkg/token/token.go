package token

type Provider interface {
	VerifyToken(token string) bool
	VerifyOnAdmin(token string) bool
}
