DROP TABLE IF EXISTS users;
CREATE TABLE IF NOT EXISTS users
(
    id    bigserial PRIMARY KEY,
    login text NOT NULL UNIQUE
);


DROP TABLE IF EXISTS tests;
CREATE TABLE IF NOT EXISTS tests
(
    id         bigserial PRIMARY KEY,
    name       text NOT NULL UNIQUE,
    title      text NOT NULL,
    topic      text NOT NULL,
    is_deleted boolean DEFAULT FALSE
);


DROP TABLE IF EXISTS commits;
CREATE TABLE IF NOT EXISTS commits
(
    id         bigserial PRIMARY KEY,
    user_id    bigint REFERENCES users (id),
    "commit"   text        NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX commits__user_id ON commits (user_id);


CREATE TYPE report_status_t AS ENUM (
    'success',
    'failure',
    'exception'
    );


DROP TABLE IF EXISTS checks;
CREATE TABLE IF NOT EXISTS checks
(
    id         bigserial PRIMARY KEY,
    user_id    bigint REFERENCES users (id),
    commit_id  bigint REFERENCES commits (id),
    test_id    bigint REFERENCES tests (id)    DEFAULT NULL,
    version_id bigint REFERENCES versions (id) DEFAULT NULL,
    name       text            NOT NULL,
    status     report_status_t NOT NULL,
    output     text            NOT NULL
);
CREATE INDEX checks__user_id ON checks (user_id);
CREATE INDEX checks__commit_id ON checks (commit_id);
CREATE UNIQUE INDEX checks__user_commit ON checks (user_id, commit_id);

CREATE TYPE run_status_t AS ENUM (
    'success',
    'failure'
    );


DROP TABLE IF EXISTS runs;
CREATE TABLE IF NOT EXISTS runs
(
    id     bigserial PRIMARY KEY,
    "hash" text         NOT NULL UNIQUE,
    status run_status_t NOT NULL,
    perf   bigint       NOT NULL
);


DROP TABLE IF EXISTS versions;
CREATE TABLE IF NOT EXISTS versions
(
    id     bigserial PRIMARY KEY,
    tag    text NOT NULL,
    active bool DEFAULT FALSE
);


DROP TABLE IF EXISTS baselines;
CREATE TABLE IF NOT EXISTS baselines
(
    id         bigserial PRIMARY KEY,
    version_id bigint REFERENCES versions (id) DEFAULT NULL,
    test_id    bigint REFERENCES tests (id)    DEFAULT NULL,
    run_id     bigint REFERENCES tests (id)    DEFAULT NULL
);


CREATE TYPE task_status_t AS ENUM (
    'enqueued',
    'executing',
    'success',
    'failure'
    );


DROP TABLE IF EXISTS tasks;
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

CREATE INDEX "tasks__enqueued_idx" ON tasks (topic, status) WHERE status = 'enqueued';
