CREATE TABLE words (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    text TEXT NOT NULL,
    language TEXT NOT NULL,
    definitions JSONB NOT NULL,
    etymology TEXT,
    translations JSONB,
    search_terms TEXT[] NOT NULL,
    lemma TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT unique_word_language UNIQUE (text, language)
);

CREATE INDEX idx_words_text ON words USING GIN (to_tsvector('english', text));
CREATE INDEX idx_words_search_terms ON words USING GIN (search_terms);
CREATE INDEX idx_words_language ON words (language);
