-- Convert definitions back to text array
ALTER TABLE words 
  ALTER COLUMN definitions TYPE TEXT[] USING 
    (SELECT array_agg(d->>'text') FROM jsonb_array_elements(definitions) d);

-- Drop the new columns
ALTER TABLE words 
  DROP COLUMN IF EXISTS word_forms,
  DROP COLUMN IF EXISTS search_terms,
  DROP COLUMN IF EXISTS lemma;

-- Add back the gender column
ALTER TABLE words ADD COLUMN IF NOT EXISTS gender TEXT;

-- Drop the search_terms index
DROP INDEX IF EXISTS idx_words_search_terms;
