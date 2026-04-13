package repository

const (
	GET_PROFILE_BY_USER_ID     = "SELECT id, username, created_at, updated_at FROM profiles WHERE id = $1"
	UPDATE_PROFILE_BY_USER_ID  = "UPDATE profiles SET username = $2, updated_at = now() WHERE id = $1 RETURNING id, username, created_at, updated_at"
	DELETE_PROFILE_BY_USER_ID  = "DELETE FROM profiles WHERE id = $1 RETURNING id"
	CHANGE_PASSWORD_BY_USER_ID = "UPDATE profiles SET password = $2 WHERE id = $1 RETURNING id, username, created_at, updated_at"
	GET_PASSWORD_BY_USER_ID    = "SELECT password FROM profiles WHERE id = $1"
)
