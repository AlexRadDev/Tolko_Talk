CREATE TABLE IF NOT EXISTS channels_lists
    (
        id SERIAL PRIMARY KEY,
        chat_id BIGINT NOT NULL,
        name_channel VARCHAR(255) NOT NULL,
        link_channel TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT NOW(),
        updated_at TIMESTAMP DEFAULT NOW(),
        UNIQUE(chat_id, name_channel)
);

-- CREATE INDEX IF NOT EXISTS idx_channels_chat_id ON channels_lists(chat_id);
-- CREATE INDEX IF NOT EXISTS idx_channels_name ON cchannels_lists(name_channel);