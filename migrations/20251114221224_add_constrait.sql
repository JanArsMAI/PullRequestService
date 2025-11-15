-- +goose Up
-- +goose StatementBegin
ALTER TABLE pull_request_reviewers
ADD CONSTRAINT unique_pr_reviewer UNIQUE (pull_request_id, reviewer_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE pull_request_reviewers
DROP CONSTRAINT unique_pr_reviewer;
-- +goose StatementEnd