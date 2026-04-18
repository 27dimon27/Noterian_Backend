package repository

const (
	CHECK_USER_EXISTS    = "SELECT EXISTS(SELECT 1 FROM profiles WHERE username = $1)"
	CREATE_USER          = "INSERT INTO profiles (id, username, password, token_version, created_at, updated_at) VALUES ($1, $2, $3, $4, now(), now())"
	GET_USER_BY_USERNAME = "SELECT id, username, password, token_version, created_at, updated_at FROM profiles WHERE username = $1"
)
