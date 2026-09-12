-- name: CreateUser :one
INSERT INTO users (username, display_name, password_hash, bio, sex, birthday, created_at, last_updated_at, client_s)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: DeleteActivitiesByUser :exec
-- Deletes every activity belonging to any of the user's sessions. Needed
-- before deleting the user's custom exercises, since activities.exercise_id
-- is ON DELETE RESTRICT (to protect shared/global exercises from accidental
-- deletion) — that RESTRICT can otherwise conflict with the CASCADE from
-- users -> exercises during a single `DELETE FROM users`.
DELETE FROM activities
WHERE session_id IN (SELECT id FROM sessions WHERE user_id = $1);

-- name: DeleteExercisesByUser :exec
-- Deletes the user's own custom exercises. Must run after
-- DeleteActivitiesByUser so no activity still references them.
DELETE FROM exercises WHERE user_id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: UpdateUserByID :one
UPDATE users
SET (username, display_name, bio, sex, birthday, last_updated_at, client_s) = ($2, $3, $4, $5, $6, $7, $8)
WHERE id = $1
RETURNING *;
