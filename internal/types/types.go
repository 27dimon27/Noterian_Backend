package types

type ctxKey int

const (
	RequestIDKey ctxKey = iota
	UserIDKey
	JWTTokenKey
)
