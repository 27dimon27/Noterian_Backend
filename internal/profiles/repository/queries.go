package repository

const (
	GET_PROFILE_BY_USER_ID = `
		SELECT id, username, created_at, updated_at 
		FROM profiles 
		WHERE id = $1
	`

	GET_PROFILE_BY_USERNAME = `
		SELECT id, username, created_at, updated_at 
		FROM profiles 
		WHERE username = $1
	`

	UPDATE_PROFILE_BY_USER_ID = `
		UPDATE profiles 
		SET username = $2, updated_at = now() 
		WHERE id = $1 
		RETURNING id, username, created_at, updated_at
	`

	DELETE_PROFILE_BY_USER_ID = `
		DELETE FROM profiles 
		WHERE id = $1 
		RETURNING id
	`

	GET_AVATAR_BY_PROFILE_ID = `
		SELECT id, profile_id, minio_key, avatar_url, url_expires_at, created_at, updated_at 
		FROM avatars 
		WHERE profile_id = $1
	`

	UPDATE_AVATAR_URL = `
		UPDATE avatars 
		SET avatar_url = $2, url_expires_at = $3, updated_at = now() 
		WHERE id = $1 
		RETURNING avatar_url, url_expires_at, updated_at
	`

	CREATE_AVATAR = `
		INSERT INTO avatars (id, profile_id, minio_key, avatar_url, url_expires_at, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, now(), now()) 
		RETURNING id, profile_id, minio_key, avatar_url, url_expires_at, created_at, updated_at
	`

	DELETE_AVATAR_BY_ID = `
		DELETE FROM avatars 
		WHERE profile_id = $1 
		RETURNING minio_key
	`

	CHANGE_PASSWORD_BY_USER_ID = `
		UPDATE profiles SET password = $2, updated_at = now() 
		WHERE id = $1 
		RETURNING id, username, created_at, updated_at
	`

	GET_PASSWORD_BY_USER_ID = `
		SELECT password 
		FROM profiles 
		WHERE id = $1
	`

	CHECK_USER_EXISTS = "SELECT EXISTS(SELECT 1 FROM profiles WHERE username = $1)"

	CREATE_USER = `
		INSERT INTO profiles (id, username, password, token_version, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, now(), now())
	`

	GET_USER_BY_USERNAME = `
		SELECT id, username, password, token_version, created_at, updated_at 
		FROM profiles 
		WHERE username = $1
	`
)
