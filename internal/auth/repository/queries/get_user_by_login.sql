SELECT id, username, password, token_version, created_at, updated_at 
FROM accounts 
WHERE username = $1