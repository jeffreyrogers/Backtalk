CREATE TABLE comments (
  id      BIGSERIAL   PRIMARY KEY,
  created timestamptz NOT NULL DEFAULT NOW(),
  slug    text        NOT NULL,
  author  text        NOT NULL,
  content text        NOT NULL,
  UNIQUE (slug, author, content)
);

CREATE TABLE users (
  id       SERIAL  PRIMARY KEY,
  email    text    UNIQUE NOT NULL,
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
