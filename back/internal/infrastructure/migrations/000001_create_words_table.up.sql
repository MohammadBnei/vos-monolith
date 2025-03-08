CREATE TABLE IF NOT EXISTS words (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    text VARCHAR(255) NOT NULL,
    language VARCHAR(10) NOT NULL,
    definitions TEXT[] NOT NULL DEFAULT '{}',
    examples TEXT[] NOT NULL DEFAULT '{}',
    pronunciation TEXT,
    etymology TEXT,
    translations JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE(text, language)
);

CREATE INDEX idx_words_text_language ON words(text, language);
CREATE INDEX idx_words_language ON words(language);
CREATE INDEX idx_words_updated_at ON words(updated_at DESC);
