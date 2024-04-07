package token

type TokenProvider interface {
	VerifyToken(token string) bool
	VerifyOnAdmin(token string) bool
}
