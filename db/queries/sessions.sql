-- name: CreateSession :one
INSERT INTO sessions (
    user_id, session_type, start_time, end_time,
    total_time, total_weight, overall_perceived_effort, burned_cals, user_notes, client_s
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: ListSessionsByUser :many
SELECT * FROM sessions WHERE user_id = $1 ORDER BY start_time DESC;

-- name: GetSessionByID :one
SELECT * FROM sessions WHERE id = $1;

-- name: CreateActivity :one
INSERT INTO activities (
    user_id_in_act, session_id, exercise_id, activity_type, reps, weight,
    sort_order, start_time, end_time, total_time, perceived_effort, client_s
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: ListActivitiesBySession :many
SELECT * FROM activities WHERE session_id = $1 ORDER BY sort_order;

-- name: ListActivitiesByUser :many
SELECT * FROM activities WHERE user_id_in_act = $1 ORDER BY session_id, sort_order;

-- name: GetMaxWeightByExerciseID :one
SELECT COALESCE(ROUND(MAX(weight)), 0)::integer
FROM activities
WHERE user_id_in_act = $1
  AND exercise_id = $2;
