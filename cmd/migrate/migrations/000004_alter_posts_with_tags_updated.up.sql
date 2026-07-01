ALTER TABLE posts
ALTER COLUMN tags TYPE VARCHAR(100) [],
    ALTER COLUMN updated_at TYPE timestamp(0) with time zone,
    ALTER COLUMN updated_at
SET NOT NULL,
    ALTER COLUMN updated_at
SET DEFAULT NOW();