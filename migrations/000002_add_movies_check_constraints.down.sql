ALTER TABLE movies
DROP CONSTRAINT IF EXISTS movies_runtime_check;

ALTER TABLE movies
DROP CONSTRAINT IF EXISTS movies_year_check;

ALTER TABLE movies
DROP CONSTRAINT IF EXISts genres_length_check;
