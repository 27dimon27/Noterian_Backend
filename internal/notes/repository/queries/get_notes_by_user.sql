SELECT id, user_id, title, parent_id, created_at, updated_at 
FROM notes 
WHERE user_id = $1 
ORDER BY updated_at DESC