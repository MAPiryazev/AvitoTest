INSERT INTO teams(team_name) VALUES ('backend'), ('payments');

INSERT INTO users(user_id, username, team_id, is_active)
VALUES 
('u1', 'name1', 1, TRUE),
('u2', 'name2', 1, TRUE),
('u3', 'name3', 1, TRUE),
('u4', 'name4', 2, TRUE),
('u5', 'name5', 2, TRUE);

INSERT INTO pull_requests(pull_request_id, pull_request_name, author_id)
VALUES 
('pr-1001', 'Add search', 1),
('pr-1002', 'bugfix', 4);

INSERT INTO pull_request_reviewers(pull_request_id, reviewer_id)
VALUES
(1, 2),
(1, 3);
