CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE OR REPLACE FUNCTION generate_prefixed_id(prefix TEXT)
RETURNS TEXT
LANGUAGE SQL
VOLATILE
AS $$
    SELECT prefix || '_' || encode(gen_random_bytes(16), 'hex');
$$;

CREATE TABLE users (
    id              TEXT PRIMARY KEY DEFAULT generate_prefixed_id('usr'),
    username        TEXT NOT NULL UNIQUE,
    display_name    TEXT NOT NULL,
    password_hash   TEXT NOT NULL,
    bio             TEXT,
    sex             TEXT,
    birthday        DATE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_updated_at TIMESTAMPTZ,
    client_s        JSONB
);

CREATE TABLE exercises (
    id                TEXT PRIMARY KEY DEFAULT generate_prefixed_id('exr'),
    user_id           TEXT REFERENCES users(id) ON DELETE CASCADE, -- NULL in global exercises
    name              TEXT NOT NULL,
    alternative_names TEXT[],
    explanation       TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    client_s          JSONB
);

CREATE INDEX idx_exercises_user_id ON exercises(user_id);

CREATE UNIQUE INDEX unique_user_exercise_name
    ON exercises(user_id, name)
    WHERE user_id IS NOT NULL;

CREATE TABLE sessions (
    id                       TEXT PRIMARY KEY DEFAULT generate_prefixed_id('ses'),
    user_id                  TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_type             TEXT NOT NULL,
    start_time               TIMESTAMPTZ NOT NULL,
    end_time                 TIMESTAMPTZ NOT NULL,
    total_time               INTERVAL NOT NULL,
    total_weight             INTEGER NOT NULL,
    overall_perceived_effort INTEGER,
    burned_cals              INTEGER,
    user_notes               TEXT,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    client_s                 JSONB,

    CONSTRAINT valid_session_effort CHECK (
        overall_perceived_effort IS NULL OR overall_perceived_effort BETWEEN 1 AND 10
    ),
    CONSTRAINT valid_session_times CHECK (end_time >= start_time)
);

CREATE INDEX idx_sessions_user_id_start_time ON sessions(user_id, start_time DESC);

CREATE TABLE activities (
    id               TEXT PRIMARY KEY DEFAULT generate_prefixed_id('act'),
    session_id       TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    exercise_id      TEXT REFERENCES exercises(id) ON DELETE RESTRICT,
    activity_type    TEXT NOT NULL,
    reps             INTEGER,
    weight           REAL, -- kg
    sort_order       INTEGER NOT NULL,
    start_time       TIMESTAMPTZ NOT NULL,
    end_time         TIMESTAMPTZ NOT NULL,
    total_time       INTERVAL NOT NULL,
    perceived_effort INTEGER,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    client_s         JSONB,

    CONSTRAINT valid_activity_type CHECK (activity_type IN ('exercise', 'rest', 'other')),
    CONSTRAINT exercise_id_matches_type CHECK (
        (activity_type = 'exercise' AND exercise_id IS NOT NULL) OR
        (activity_type = 'rest'     AND exercise_id IS NULL) OR
        activity_type = 'other'
    ),
    CONSTRAINT valid_activity_effort CHECK (
        (activity_type = 'exercise' AND perceived_effort IS NULL OR perceived_effort BETWEEN 1 AND 10) OR
        (activity_type = 'rest'     AND perceived_effort IS NULL) OR
        activity_type = 'other'
    ),
    CONSTRAINT valid_activity_times CHECK (end_time >= start_time),
    CONSTRAINT unique_session_sort_order UNIQUE (session_id, sort_order),
    CONSTRAINT valid_activity_reps CHECK (
        (activity_type = 'exercise' AND reps IS NOT NULL) OR
        (activity_type = 'rest' AND reps IS NULL) OR
        activity_type = 'other'
    ),
    CONSTRAINT valid_activity_weight CHECK (
        (activity_type = 'exercise' AND weight IS NOT NULL) OR
        (activity_type = 'rest' AND weight IS NULL) OR
        activity_type = 'other'
    )
);

CREATE INDEX idx_activities_session_id ON activities(session_id);
CREATE INDEX idx_activities_exercise_id ON activities(exercise_id);

CREATE TABLE auth_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_auth_tokens_user_id ON auth_tokens(user_id);
