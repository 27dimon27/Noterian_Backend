package repository

const (
	CREATE_ATTACHMENT = `
		INSERT INTO attachments (id, block_id, file_name, file_size, mime_type, minio_key, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
		RETURNING id, block_id, file_name, file_size, mime_type, minio_key, created_at, updated_at
	`

	GET_ATTACHMENT_BY_BLOCK_ID = `
		SELECT id, block_id, file_name, file_size, mime_type, minio_key, created_at, updated_at 
		FROM attachments 
		WHERE block_id = $1
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
