CREATE TABLE users (
  id       SERIAL  PRIMARY KEY,
  email    text    NOT NULL UNIQUE,
  hash     bytea   NOT NULL,
  salt     bytea   NOT NULL,
  is_admin boolean NOT NULL DEFAULT false
);

CREATE TABLE sessions (
  session_id  bytea       PRIMARY KEY,
  uid         serial      NOT NULL REFERENCES users,
  login_time  timestamptz NOT NULL DEFAULT NOW(),
  last_seen   timestamptz NOT NULL DEFAULT NOW()
);
