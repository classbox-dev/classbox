CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- -----------------------------------------------------------------------------

DROP TABLE IF EXISTS users CASCADE;
CREATE TABLE IF NOT EXISTS users
(
    id              bigserial PRIMARY KEY,
    github_id       bigint NOT NULL UNIQUE,
    login           text   NOT NULL,
    email           text   NOT NULL,
    repository_id   bigint NOT NULL,
    repository_name text   NOT NULL,
    installation_id bigint,
    honor_code      boolean DEFAULT FALSE
);

-- -----------------------------------------------------------------------------

DROP TABLE IF EXISTS sessions;
CREATE TABLE IF NOT EXISTS sessions
(
    id         bigserial PRIMARY KEY,
    user_id    bigserial REFERENCES users (id),
    session    text        NOT NULL,
    expires_at timestamptz NOT NULL
);

CREATE INDEX sessions__user_id ON sessions (user_id);
CREATE INDEX sessions__session ON sessions (session);

-- -----------------------------------------------------------------------------

DROP TABLE IF EXISTS tests CASCADE;
CREATE TABLE IF NOT EXISTS tests
(
    id          bigserial PRIMARY KEY,
    name        text   NOT NULL UNIQUE,
    description text   NOT NULL,
    topic       text   NOT NULL,
    score       bigint NOT NULL,
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
    id          bigserial PRIMARY KEY,
    "hash"      text         NOT NULL UNIQUE,
    status      run_status_t NOT NULL,
    output      text         NOT NULL,
    score       bigint       NOT NULL,
    test_id     bigint REFERENCES tests (id) DEFAULT NULL,
    is_baseline boolean                      DEFAULT FALSE
);
CREATE INDEX runs_hash_baseline ON runs ("hash", is_baseline);

-- -----------------------------------------------------------------------------

DROP TABLE IF EXISTS commits CASCADE;
CREATE TABLE IF NOT EXISTS commits
(
    id           bigserial PRIMARY KEY,
    user_id      bigint REFERENCES users (id),
    commit       text   NOT NULL,
    check_run_id bigint NOT NULL,
    is_checked   boolean DEFAULT FALSE
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
    commit_id bigint REFERENCES commits (id),
    test_id   bigint REFERENCES tests (id) DEFAULT NULL,
    name      text           NOT NULL,
    status    check_status_t NOT NULL,
    output    text           NOT NULL,
    run_id    bigint REFERENCES runs (id)  DEFAULT NULL,
    is_cached boolean                      DEFAULT FALSE
);
CREATE INDEX checks__commit_id ON checks (commit_id);
CREATE INDEX checks__test_id ON checks (test_id) WHERE test_id IS NOT NULL;
CREATE INDEX checks__run_id ON checks (run_id) WHERE run_id IS NOT NULL;

-- -----------------------------------------------------------------------------

DROP TYPE IF EXISTS task_status_t CASCADE;
CREATE TYPE task_status_t AS ENUM (
    'enqueued',
    'executing',
    'finished'
    );

DROP TABLE IF EXISTS tasks CASCADE;
CREATE TABLE IF NOT EXISTS tasks
(
    id          uuid PRIMARY KEY               DEFAULT uuid_generate_v4(),
    enqueued_at timestamptz NOT NULL           DEFAULT CURRENT_TIMESTAMP,
    started_at  timestamptz,
    finished_at timestamptz,
    commit_id   bigint REFERENCES commits (id) DEFAULT NULL UNIQUE,
    status      task_status_t                  DEFAULT 'enqueued'
);
CREATE INDEX "tasks__enqueued_idx" ON tasks (status) WHERE status = 'enqueued';

-- -----------------------------------------------------------------------------

DROP TABLE IF EXISTS courses CASCADE;
CREATE TABLE IF NOT EXISTS courses
(
    id         bigserial PRIMARY KEY,
    name       text        NOT NULL UNIQUE,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_ready   boolean              DEFAULT FALSE
);

INSERT INTO courses (name)
VALUES ('stdlib');
