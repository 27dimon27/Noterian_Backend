package repository

const (
	GET_BLOCKS_BY_NOTE = "SELECT b.id, b.note_id, b.block_type_id, b.position, b.content, bs.id, bs.formatting, bs.created_at, bs.updated_at FROM blocks b LEFT JOIN block_states bs ON b.id = bs.block_id WHERE b.note_id = $1 ORDER BY b.position, bs.created_at"
	GET_NOTE_BY_ID     = "SELECT id, user_id, title, parent_id, created_at, updated_at FROM notes WHERE id = $1"
	GET_NOTES_BY_USER  = "SELECT id, user_id, title, parent_id, created_at, updated_at FROM notes WHERE user_id = $1 ORDER BY updated_at DESC"
	CREATE_NOTE        = "INSERT INTO notes (user_id, title, parent_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id, user_id, title, parent_id, created_at, updated_at"
	UPDATE_NOTE        = "UPDATE notes SET title = $2, parent_id = $3, updated_at = $4 WHERE id = $1 RETURNING id, user_id, title, parent_id, created_at, updated_at"
	DELETE_NOTE        = "DELETE FROM notes WHERE id = $1 RETURNING id"
)
