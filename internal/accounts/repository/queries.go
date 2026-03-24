package repository

const (
	GET_ACCOUNT_BY_USER_ID = "SELECT id, username FROM accounts WHERE id = $1"
)
