-- Teams
CREATE TABLE teams (
    id SERIAL PRIMARY KEY,
    team_name VARCHAR(100) UNIQUE NOT NULL
);

-- Users
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(50) UNIQUE NOT NULL,     
    username VARCHAR(100) NOT NULL,
    team_id INT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

-- Pull Requests
CREATE TABLE pull_requests (
    id SERIAL PRIMARY KEY,
    pull_request_id VARCHAR(50) UNIQUE NOT NULL,
    pull_request_name VARCHAR(255) NOT NULL,
    author_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(10) NOT NULL DEFAULT 'OPEN', -- OPEN или MERGED
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    merged_at TIMESTAMP
);

-- Pull Request Reviewers
CREATE TABLE pull_request_reviewers (
    pull_request_id INT NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
    reviewer_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (pull_request_id, reviewer_id)
);



CREATE UNIQUE INDEX idx_users_user_id ON users(user_id);
CREATE UNIQUE INDEX idx_pr_pr_id ON pull_requests(pull_request_id);