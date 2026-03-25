package repository

const (
	GET_PROFILE_BY_USER_ID = "SELECT id, username FROM accounts WHERE id = $1"
)
