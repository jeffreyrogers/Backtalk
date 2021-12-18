CREATE TABLE comments (
  id      BIGSERIAL   PRIMARY KEY,
  created timestamptz NOT NULL DEFAULT NOW(),
  slug    text        NOT NULL,
  author  text        NOT NULL,
  content text        NOT NULL,
  UNIQUE (slug, author, content)
);
