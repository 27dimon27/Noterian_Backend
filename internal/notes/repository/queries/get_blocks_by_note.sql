SELECT 
    b.id, b.note_id, b.block_type_id, b.position, b.content,
    bs.id, bs.formatting, bs.created_at, bs.updated_at
FROM blocks b
LEFT JOIN block_states bs ON b.id = bs.block_id
WHERE b.note_id = $1
ORDER BY b.position, bs.created_at