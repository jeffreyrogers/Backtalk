CREATE TABLE users (
  id       SERIAL  PRIMARY KEY,
  email    text    NOT NULL UNIQUE,
  hash     text    NOT NULL,
  salt     text    NOT NULL,
  is_admin boolean NOT NULL DEFAULT false
);

CREATE TABLE sessions (
  session_key text        PRIMARY KEY,
  uid         serial      NOT NULL REFERENCES users,
  login_time  timestamptz NOT NULL,
  last_seen   timestamptz NOT NULL
);
