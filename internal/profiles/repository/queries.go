package repository

const (
	GET_PROFILE_BY_USER_ID    = "SELECT id, username, created_at, updated_at FROM profiles WHERE id = $1"
	GET_PROFILE_BY_USERNAME   = "SELECT id, username, created_at, updated_at FROM profiles WHERE username = $1"
	UPDATE_PROFILE_BY_USER_ID = "UPDATE profiles SET username = $2, updated_at = now() WHERE id = $1 RETURNING id, username, created_at, updated_at"
	DELETE_PROFILE_BY_USER_ID = "DELETE FROM profiles WHERE id = $1 RETURNING id"
	GET_AVATAR_BY_PROFILE_ID  = `
		SELECT id, profile_id, minio_key, avatar_url, url_expires_at, created_at, updated_at 
		FROM avatars 
		WHERE profile_id = $1
	`
	UPDATE_AVATAR_URL = `
		UPDATE avatars 
		SET avatar_url = $1, url_expires_at = $2, updated_at = $3
		WHERE id = $4
		RETURNING avatar_url, url_expires_at, updated_at
	`
	CREATE_AVATAR = `
		INSERT INTO avatars (id, profile_id, minio_key, avatar_url, url_expires_at, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) 
		RETURNING id, profile_id, minio_key, avatar_url, url_expires_at, created_at, updated_at
	`
	DELETE_AVATAR_BY_ID = `
		DELETE FROM avatars 
		WHERE profile_id = $1 
		RETURNING minio_key
	`
	CHANGE_PASSWORD_BY_USER_ID = "UPDATE profiles SET password = $2 WHERE id = $1 RETURNING id, username, created_at, updated_at"
	GET_PASSWORD_BY_USER_ID    = "SELECT password FROM profiles WHERE id = $1"
)
