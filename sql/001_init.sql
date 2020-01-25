DROP TABLE IF EXISTS users CASCADE;
CREATE TABLE IF NOT EXISTS users
(
    id              bigserial PRIMARY KEY,
    github_id       bigint NOT NULL UNIQUE,
    login           text   NOT NULL,
    email           text   NOT NULL,
    repository_id   bigint NOT NULL,
    repository_name text   NOT NULL,
    installation_id bigint
);

-- -----------------------------------------------------------------------------

DROP TABLE IF EXISTS tests CASCADE;
CREATE TABLE IF NOT EXISTS tests
(
    id          bigserial PRIMARY KEY,
    name        text NOT NULL UNIQUE,
    description text NOT NULL,
    topic       text NOT NULL,
    is_deleted  boolean DEFAULT FALSE
);

-- -----------------------------------------------------------------------------

DROP TYPE IF EXISTS run_status_t CASCADE;
CREATE TYPE run_status_t AS ENUM (
    'success',
    'failure'
    );

DROP TABLE IF EXISTS runs CASCADE;
CREATE TABLE IF NOT EXISTS runs
(
    id     bigserial PRIMARY KEY,
    "hash" text         NOT NULL UNIQUE,
    status run_status_t NOT NULL,
    output text         NOT NULL,
    score  bigint       NOT NULL
);

-- -----------------------------------------------------------------------------

DROP TABLE IF EXISTS baselines CASCADE;
CREATE TABLE IF NOT EXISTS baselines
(
    id      bigserial PRIMARY KEY,
    test_id bigint REFERENCES tests (id) DEFAULT NULL,
    run_id  bigint REFERENCES runs (id)  DEFAULT NULL
);
CREATE INDEX commits__test_id ON baselines (test_id);
CREATE INDEX commits__run_id ON baselines (run_id);

-- -----------------------------------------------------------------------------

DROP TABLE IF EXISTS commits CASCADE;
CREATE TABLE IF NOT EXISTS commits
(
    id         bigserial PRIMARY KEY,
    user_id    bigint REFERENCES users (id),
    commit     text        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX commits__user_id ON commits (user_id);
CREATE UNIQUE INDEX commits__user_commit ON commits (user_id, commit);

-- -----------------------------------------------------------------------------

DROP TYPE IF EXISTS check_status_t CASCADE;
CREATE TYPE check_status_t AS ENUM (
    'success',
    'failure',
    'exception'
    );

DROP TABLE IF EXISTS checks CASCADE;
CREATE TABLE IF NOT EXISTS checks
(
    id        bigserial PRIMARY KEY,
    user_id   bigint REFERENCES users (id),
    commit_id bigint REFERENCES commits (id),
    test_id   bigint REFERENCES tests (id) DEFAULT NULL,
    name      text           NOT NULL,
    status    check_status_t NOT NULL,
    output    text           NOT NULL,
    is_cached boolean                      DEFAULT FALSE
);
CREATE INDEX checks__user_id ON checks (user_id);
CREATE INDEX checks__commit_id ON checks (commit_id);
CREATE INDEX checks__test_id ON checks (test_id);
CREATE UNIQUE INDEX checks__user_commit ON checks (user_id, commit_id);

-- -----------------------------------------------------------------------------

DROP TYPE IF EXISTS task_status_t CASCADE;
CREATE TYPE task_status_t AS ENUM (
    'enqueued',
    'executing',
    'success',
    'failure'
    );

DROP TABLE IF EXISTS tasks CASCADE;
CREATE TABLE IF NOT EXISTS tasks
(
    id          bigserial PRIMARY KEY,
    enqueued_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at  timestamptz,
    ended_at    timestamptz,
    topic       text        NOT NULL,
    payload     jsonb       NOT NULL,
    status      task_status_t        DEFAULT 'enqueued'
);
CREATE INDEX "tasks__enqueued_idx"
    ON tasks (topic, status) WHERE status = 'enqueued';

-- -----------------------------------------------------------------------------

DROP TABLE IF EXISTS versions CASCADE;
CREATE TABLE IF NOT EXISTS versions
(
    id  bigserial PRIMARY KEY,
    tag text NOT NULL
);
