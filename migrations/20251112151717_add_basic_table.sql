-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS teams (
    id SERIAL PRIMARY KEY,
    team_name VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(50) NOT NULL,
    username VARCHAR(100) NOT NULL,
    team_id INT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    PRIMARY KEY (user_id)
);

CREATE TABLE IF NOT EXISTS pull_requests (
    pull_request_id VARCHAR(50) PRIMARY KEY,
    pull_request_name VARCHAR(200) NOT NULL,
    author_id VARCHAR(50) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    status VARCHAR(10) NOT NULL CHECK (status IN ('OPEN','MERGED')),
    need_more_reviewers BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    merged_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS pull_request_reviewers (
    pull_request_id VARCHAR(50) NOT NULL REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    reviewer_id VARCHAR(50) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (pull_request_id, reviewer_id)
);

CREATE INDEX idx_users_team ON users(team_id);
CREATE INDEX idx_pr_reviewers_user ON pull_request_reviewers(reviewer_id);
CREATE INDEX idx_pr_status ON pull_requests(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS pull_request_reviewers;
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;
-- +goose StatementEnd
