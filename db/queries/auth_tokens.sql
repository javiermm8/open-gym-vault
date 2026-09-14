-- name: CreateAuthToken :one
INSERT INTO auth_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetAuthTokenByHash :one
SELECT * FROM auth_tokens WHERE token_hash = $1;

-- name: RefreshAuthTokenExpiry :exec
UPDATE auth_tokens SET expires_at = $2 WHERE id = $1;

-- name: DeleteAuthToken :exec
DELETE FROM auth_tokens WHERE token_hash = $1;

-- name: DeleteExpiredAuthTokens :exec
DELETE FROM auth_tokens WHERE expires_at < now();

-- name: DeleteAuthTokenByUserID :exec
DELETE FROM auth_tokens WHERE user_id = $1;
