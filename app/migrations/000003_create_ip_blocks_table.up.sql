CREATE TABLE ip_blocks (
    id UUID PRIMARY KEY,
    ip INET NOT NULL,
    block_type VARCHAR(20) NOT NULL,
    reason VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NULL,
    created_by UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    comment TEXT NOT NULL DEFAULT '',

    CONSTRAINT ip_blocks_ip_idx UNIQUE (ip)
);

CREATE INDEX ip_blocks_expires_at_idx ON ip_blocks (expires_at); 