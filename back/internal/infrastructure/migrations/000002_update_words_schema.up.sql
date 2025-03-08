-- Add new columns for enhanced word structure
ALTER TABLE words 
  ADD COLUMN IF NOT EXISTS word_type TEXT,
  ADD COLUMN IF NOT EXISTS forms JSONB DEFAULT '[]'::jsonb,
  ADD COLUMN IF NOT EXISTS search_terms TEXT[] DEFAULT ARRAY[]::TEXT[],
  ADD COLUMN IF NOT EXISTS lemma TEXT;

-- Drop the gender column as it's replaced by forms
ALTER TABLE words DROP COLUMN IF EXISTS gender;

-- Create an index on search_terms for efficient searching
CREATE INDEX IF NOT EXISTS idx_words_search_terms ON words USING GIN (search_terms);

-- Update definitions column to use JSONB for storing type information
ALTER TABLE words 
  ALTER COLUMN definitions TYPE JSONB USING 
    (SELECT COALESCE(
      array_to_json(ARRAY(SELECT json_build_object('text', d, 'type', '') FROM unnest(definitions) d)),
      '[]'::jsonb
    ));

COMMENT ON COLUMN words.definitions IS 'Word definitions with type information';
COMMENT ON COLUMN words.forms IS 'Different forms of the word with attributes';
COMMENT ON COLUMN words.search_terms IS 'All searchable forms of the word';
COMMENT ON COLUMN words.lemma IS 'Base form of the word';
