-- name: ListGlobalExercises :many
SELECT * FROM exercises WHERE user_id IS NULL ORDER BY name;

-- name: ListExercisesForUser :many
-- Global exercises plus this user's own custom ones.
SELECT * FROM exercises
WHERE user_id IS NULL OR user_id = $1
ORDER BY name;

-- name: GetExerciseByID :one
SELECT * FROM exercises WHERE id = $1;

-- name: CreateGlobalExercise :one
INSERT INTO exercises (name, alternative_names, explanation)
VALUES ($1, $2, $3)
RETURNING *;

-- name: CreateCustomExercise :one
INSERT INTO exercises (user_id, name, alternative_names, explanation)
VALUES ($1, $2, $3, $4)
RETURNING *;
