package repository

const (
	GET_BLOCKS_BY_NOTE = `
		SELECT id, note_id, block_type_id, position, content, created_at, updated_at 
		FROM blocks 
		WHERE note_id = $1 
		ORDER BY position
	`

	GET_NOTE_BY_ID = `
		SELECT id, user_id, title, parent_id, is_public, is_favorite, icon, created_at, updated_at 
		FROM notes 
		WHERE id = $1
	`

	GET_NOTES_BY_USER = `
		SELECT id, user_id, title, parent_id, is_public, is_favorite, icon, created_at, updated_at 
		FROM notes 
		WHERE user_id = $1 
		ORDER BY updated_at DESC
	`

	CREATE_NOTE = `
		INSERT INTO notes (user_id, title, parent_id, is_public, is_favorite, icon, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, now(), now()) 
		RETURNING id, user_id, title, parent_id, is_public, is_favorite, icon, created_at, updated_at
	`

	UPDATE_NOTE = `
		UPDATE notes SET title = $2, parent_id = $3, is_public = $4, is_favorite = $5, icon = $6, updated_at = now() 
		WHERE id = $1 
		RETURNING id, user_id, title, parent_id, is_public, is_favorite, icon, created_at, updated_at
	`

	DELETE_NOTE = `
		DELETE FROM notes 
		WHERE id = $1 
		RETURNING id
	`

	CREATE_BLOCK = `
		INSERT INTO blocks (note_id, block_type_id, position, content, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, now(), now()) 
		RETURNING id, note_id, block_type_id, position, content, created_at, updated_at
	`

	GET_BLOCK_BY_ID = `
		SELECT id, note_id, block_type_id, position, content, created_at, updated_at 
		FROM blocks 
		WHERE id = $1
	`

	UPDATE_BLOCK_CONTENT = `
		UPDATE blocks SET content = $2, updated_at = now() 
		WHERE id = $1 
		RETURNING id, note_id, block_type_id, position, content, created_at, updated_at
	`

	UPDATE_BLOCK_POSITION = `
		UPDATE blocks SET position = $2, updated_at = now() 
		WHERE id = $1 
		RETURNING id, note_id, block_type_id, position, content, created_at, updated_at
	`

	DELETE_BLOCK = `
		DELETE FROM blocks 
		WHERE id = $1 
		RETURNING id, note_id
	`

	UPDATE_BLOCKS_POSITION_DOWN = `
		UPDATE blocks SET position = position - 1, updated_at = now() 
		WHERE note_id = $1 AND position > $2 AND position <= $3
	`

	UPDATE_BLOCKS_POSITION_UP = `
		UPDATE blocks SET position = position + 1, updated_at = now() 
		WHERE note_id = $1 AND position < $2 AND position >= $3
	`

	UPDATE_ALL_BLOCKS_POSITION_DOWN = `
		UPDATE blocks SET position = position - 1, updated_at = now() 
		WHERE note_id = $1 AND position > $2
	`

	UPDATE_ALL_BLOCKS_POSITION_UP = `
		UPDATE blocks SET position = position + 1, updated_at = now() 
		WHERE note_id = $1 AND position >= $2
	`

	GET_BLOCK_FORMATTING = `
		SELECT start_pos, end_pos, bold, italic, underline, text_align 
		FROM block_formatting 
		WHERE block_id = $1 
		ORDER BY start_pos
	`

	GET_BLOCKS_FORMATTING = `
		SELECT block_id, start_pos, end_pos, bold, italic, underline, text_align 
		FROM block_formatting 
		WHERE block_id = ANY($1::UUID[]) 
		ORDER BY block_id, start_pos
	`

	INSERT_BLOCK_FORMATTING = `
		INSERT INTO block_formatting (block_id, start_pos, end_pos, bold, italic, underline, text_align, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now())
	`

	DELETE_BLOCK_FORMATTING = `
		DELETE FROM block_formatting 
		WHERE block_id = $1
		RETURNING block_id
	`

	GET_SUBNOTES_BY_NOTE = `
		SELECT id, user_id, title, parent_id, is_public, is_favorite, icon, created_at, updated_at 
		FROM notes 
		WHERE parent_id = $1 
		ORDER BY updated_at DESC
	`
)
