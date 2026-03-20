SELECT id, user_id, title, parent_id, created_at, updated_at 
FROM notes 
WHERE id = $1