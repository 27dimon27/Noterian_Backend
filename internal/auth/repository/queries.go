package repository

const (
	CHECK_USER_EXISTS = "SELECT EXISTS(SELECT 1 FROM accounts WHERE username = $1)"
	CREATE_USER       = "INSERT INTO accounts (id, username, password, token_version, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)"
	GET_USER_BY_LOGIN = "SELECT id, username, password, token_version, created_at, updated_at FROM accounts WHERE username = $1"
)
