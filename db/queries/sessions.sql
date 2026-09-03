-- name: CreateSession :one
INSERT INTO sessions (
    user_id, session_type, start_time, end_time,
    total_time, total_weight, overall_perceived_effort, burned_cals, user_notes
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: ListSessionsByUser :many
SELECT * FROM sessions WHERE user_id = $1 ORDER BY start_time DESC;

-- name: GetSessionByID :one
SELECT * FROM sessions WHERE id = $1;

-- name: CreateActivity :one
INSERT INTO activities (
    session_id, exercise_id, activity_type, reps, weight,
    sort_order, start_time, end_time, total_time, perceived_effort
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: ListActivitiesBySession :many
SELECT * FROM activities WHERE session_id = $1 ORDER BY sort_order;
