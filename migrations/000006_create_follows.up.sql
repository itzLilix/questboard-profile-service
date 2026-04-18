CREATE TABLE IF NOT EXISTS follows (
    follower_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    followed_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (follower_id, followed_id),
    CHECK (follower_id <> followed_id)
);

CREATE INDEX idx_follows_followed ON follows(followed_id);