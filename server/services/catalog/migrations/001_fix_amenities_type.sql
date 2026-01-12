-- Fix amenities column type mismatch (text[] -> jsonb)

-- Drop existing column and recreate as JSONB (data loss acceptable for dev environment, or cast if needed)
-- Since it's text[] in DB but code sends JSON, we should just drop and re-add or cast. 
-- Casting text[] to jsonb is complex: to_jsonb(amenities) might work.

ALTER TABLE stations 
    ALTER COLUMN amenities TYPE JSONB USING to_jsonb(amenities);

-- If existing data is malformed (e.g. empty), default to empty array
ALTER TABLE stations ALTER COLUMN amenities SET DEFAULT '[]'::jsonb;
