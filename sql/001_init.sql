DROP TABLE IF EXISTS users;
CREATE TABLE IF NOT EXISTS users
(
    id            bigserial PRIMARY KEY,
    login         text NOT NULL UNIQUE,
    full_name     text NOT NULL,
    auth_token    text NOT NULL,
    refresh_token text NOT NULL,
    uuid          TEXT NOT NULL
);


DROP TABLE IF EXISTS user_cookies;
CREATE TABLE IF NOT EXISTS user_cookies
(
    id        bigserial PRIMARY KEY,
    user_id   bigserial REFERENCES users (id),
    cookie    text        NOT NULL,
    expire_at timestamptz NOT NULL
);

CREATE INDEX user_cookies__user_id ON user_cookies (user_id);
CREATE INDEX user_cookies__cookie ON user_cookies (cookie);


DROP TABLE IF EXISTS problems;
CREATE TABLE IF NOT EXISTS problems
(
    id     bigserial PRIMARY KEY,
    name   text   NOT NULL UNIQUE,
    title  text   NOT NULL,
    score  bigint NOT NULL,
    cycles bigint NOT NULL,
    files  text[]
);

DROP TABLE IF EXISTS submissions;
CREATE TABLE IF NOT EXISTS submissions
(
    id         bigserial PRIMARY KEY,
    user_id    bigint REFERENCES users (id),
    problem_id bigint REFERENCES problems (id),
    fhash      text        NOT NULL,
    "commit"   text        NOT NULL,
    is_passed  boolean              DEFAULT FALSE,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX submissions__user_id ON submissions (user_id);

CREATE TYPE status_t AS ENUM (
    'enqueued',
    'executing',
    'success',
    'failed'
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
    status      status_t             DEFAULT 'enqueued'
);

CREATE INDEX "tasks__enqueued_idx" ON tasks (topic, status) WHERE status = 'enqueued';
