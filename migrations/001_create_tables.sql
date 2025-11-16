CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(50) PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    team_name VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT true
);

CREATE TABLE IF NOT EXISTS teams (
    team_name VARCHAR(100) PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS pull_requests (
    pull_request_id VARCHAR(50) PRIMARY KEY,
    pull_request_name VARCHAR(200) NOT NULL,
    author_id VARCHAR(50) NOT NULL,
    status VARCHAR(20) DEFAULT 'OPEN',
    assigned_reviewers JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    merged_at TIMESTAMP NULL
);