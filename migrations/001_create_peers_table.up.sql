CREATE TABLE IF NOT EXISTS peers (
    id VARCHAR(255) PRIMARY KEY,
    ip VARCHAR(50) NOT NULL,
    port INT NOT NULL,
    last_heartbeat TIMESTAMP DEFAULT NOW(),
    device_capabilities JSONB
);
