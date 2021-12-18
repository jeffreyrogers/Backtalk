ALTER TABLE comments
ADD CONSTRAINT unique_comments UNIQUE(slug, author, content);
