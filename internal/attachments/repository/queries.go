package repository

const (
	CREATE_ATTACHMENT = `
		INSERT INTO attachments (id, block_id, minio_key, attach_url, url_expires_at, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, now(), now()) 
		RETURNING id, block_id, minio_key, attach_url, url_expires_at, created_at, updated_at
	`

	GET_ATTACHMENT_BY_BLOCK_ID = `
		SELECT id, block_id, minio_key, attach_url, url_expires_at, created_at, updated_at 
		FROM attachments 
		WHERE block_id = $1
	`

	UPDATE_ATTACHMENT_URL = `
		UPDATE attachments 
		SET attach_url = $2, url_expires_at = $3, updated_at = now()
		WHERE id = $1
		RETURNING attach_url, url_expires_at, updated_at
	`

	DELETE_ATTACHMENT_BY_ID = `
		DELETE FROM attachments 
		WHERE block_id = $1 
		RETURNING minio_key
	`

	GET_NOTE_BY_ID = `
		SELECT id, user_id, title, parent_id, created_at, updated_at 
		FROM notes 
		WHERE id = $1
	`

	GET_BLOCK_BY_ID = `
		SELECT id, note_id, block_type_id, position, content, created_at, updated_at 
		FROM blocks 
		WHERE id = $1
	`
)
