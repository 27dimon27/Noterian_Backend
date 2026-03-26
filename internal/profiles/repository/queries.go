package repository

const (
	GET_PROFILE_BY_USER_ID = "SELECT id, username FROM profiles WHERE id = $1"
)
