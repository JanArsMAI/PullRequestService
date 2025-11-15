package repos

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	entityPr "github.com/JanArsMAI/PullRequestService/internal/domain/pullrequest"
	entityTeam "github.com/JanArsMAI/PullRequestService/internal/domain/team"
	entityUser "github.com/JanArsMAI/PullRequestService/internal/domain/user"
	"github.com/JanArsMAI/PullRequestService/internal/infrastructure/repos/dto"
	"github.com/jmoiron/sqlx"
)

var (
	ErrNoUserWithId = errors.New("No user with this ID")
	ErrTeamNotFound = errors.New("team with this Name not found")
	ErrPrNotFound   = errors.New("Pull request with this id is not found")
)

type PostgresRepo struct {
	db *sqlx.DB
}

func NewPostgresRepo(db *sqlx.DB) *PostgresRepo {
	return &PostgresRepo{
		db: db,
	}
}

func (p *PostgresRepo) AddTeam(ctx context.Context, name string, users []entityUser.User) error {
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	var teamID int
	queryTeam := `INSERT INTO teams (team_name)
		VALUES ($1)
		ON CONFLICT (team_name) DO NOTHING
		RETURNING id;`
	if err := tx.QueryRowContext(ctx, queryTeam, name).Scan(&teamID); err != nil {
		if err == sql.ErrNoRows {
			if err := tx.GetContext(ctx, &teamID, `SELECT id FROM teams WHERE team_name=$1`, name); err != nil {
				tx.Rollback()
				return fmt.Errorf("error getting existing team id: %w", err)
			}
		} else {
			tx.Rollback()
			return fmt.Errorf("error inserting team: %w", err)
		}
	}
	for _, user := range users {

		if err := addUserTx(ctx, tx, user, teamID); err != nil {
			tx.Rollback()
			return fmt.Errorf("error inserting or updating user %s: %w", user.Id, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func addUserTx(ctx context.Context, tx *sqlx.Tx, user entityUser.User, teamID int) error {
	query := `INSERT INTO users (user_id, username, team_id, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE
		SET username = EXCLUDED.username,
		    is_active = EXCLUDED.is_active,
		    team_id = EXCLUDED.team_id;`
	_, err := tx.ExecContext(ctx, query, user.Id, user.Name, teamID, user.IsActive)
	if err != nil {
		return fmt.Errorf("error adding/updating user: %w", err)
	}
	return nil
}

func (p *PostgresRepo) GetTeam(ctx context.Context, id int) (*entityTeam.Team, error) {
	var team dto.TeamDto
	if err := p.db.GetContext(ctx, &team, `SELECT id, team_name FROM teams WHERE id = $1`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTeamNotFound
		}
		return nil, err
	}
	var users []dto.UserDto
	if err := p.db.SelectContext(ctx, &users, `SELECT user_id, username, is_active FROM users WHERE team_id = $1`, id); err != nil {
		return nil, err
	}
	entityTeam := &entityTeam.Team{
		Id:   team.Id,
		Name: team.Name,
	}
	for _, u := range users {
		entityTeam.Users = append(entityTeam.Users, entityUser.User{
			Id:       u.Id,
			Name:     u.Name,
			IsActive: u.IsActive,
		})
	}

	return entityTeam, nil
}

func (p *PostgresRepo) GetTeamByName(ctx context.Context, name string) (*entityTeam.Team, error) {
	var team dto.TeamDto
	if err := p.db.GetContext(ctx, &team, `SELECT id, team_name FROM teams WHERE team_name = $1`, name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTeamNotFound
		}
		return nil, fmt.Errorf("failed to get team by name %s: %w", name, err)
	}

	var users []dto.UserDto
	if err := p.db.SelectContext(ctx, &users, `SELECT user_id, username, is_active FROM users WHERE team_id = $1`, team.Id); err != nil {
		return nil, fmt.Errorf("failed to get users for team %s: %w", name, err)
	}

	entityTeam := &entityTeam.Team{
		Id:   team.Id,
		Name: team.Name,
	}

	for _, u := range users {
		entityTeam.Users = append(entityTeam.Users, entityUser.User{
			Id:       u.Id,
			Name:     u.Name,
			IsActive: u.IsActive,
		})
	}

	return entityTeam, nil
}

func (p *PostgresRepo) AddPR(ctx context.Context, pr entityPr.PullRequest) error {
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()
	queryPR := `
		INSERT INTO pull_requests (
			pull_request_id, pull_request_name, author_id, status, need_more_reviewers, created_at, merged_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err = tx.ExecContext(ctx, queryPR,
		pr.Id,
		pr.Name,
		pr.Author.Id,
		pr.Status,
		pr.NeedMoreReviewers,
		pr.CreatedAt,
		pr.MergedAt,
	)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("error inserting pull request: %w", err)
	}
	if len(pr.Reviewers) > 0 {
		queryReviewer := `INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
			VALUES ($1, $2)`
		for _, reviewer := range pr.Reviewers {
			if _, err := tx.ExecContext(ctx, queryReviewer, pr.Id, reviewer.Id); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("error inserting reviewer %s: %w", reviewer.Id, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func (p *PostgresRepo) GetPr(ctx context.Context, prID string) (*entityPr.PullRequest, error) {
	var prDto dto.PullRequestDto
	queryPR := `SELECT pull_request_id, pull_request_name, author_id, status, need_more_reviewers, created_at, merged_at
        FROM pull_requests
        WHERE pull_request_id = $1`
	if err := p.db.GetContext(ctx, &prDto, queryPR, prID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPrNotFound
		}
		return nil, fmt.Errorf("error getting pull request: %w", err)
	}
	var reviewerIDs []string
	queryReviewers := `
        SELECT reviewer_id
        FROM pull_request_reviewers
        WHERE pull_request_id = $1
    `
	if err := p.db.SelectContext(ctx, &reviewerIDs, queryReviewers, prID); err != nil {
		return nil, fmt.Errorf("get reviewers: %w", err)
	}
	pr := &entityPr.PullRequest{
		Id:                prDto.ID,
		Name:              prDto.Name,
		Author:            entityUser.User{Id: prDto.AuthorID},
		Status:            prDto.Status,
		NeedMoreReviewers: prDto.NeedMoreReviewers,
		CreatedAt:         prDto.CreatedAt,
		MergedAt:          prDto.MergedAt,
	}
	for _, rid := range reviewerIDs {
		pr.Reviewers = append(pr.Reviewers, entityUser.User{Id: rid})
	}
	return pr, nil
}

func (p *PostgresRepo) UpdatePr(ctx context.Context, prId string, newPr entityPr.PullRequest) error {
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	queryUpdate := `UPDATE pull_requests
        SET status = $1,
            need_more_reviewers = $2,
            merged_at = $3
        WHERE pull_request_id = $4`
	_, err = tx.ExecContext(ctx, queryUpdate, newPr.Status, newPr.NeedMoreReviewers, newPr.MergedAt, prId)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error updating pull_request: %w", err)
	}
	_, err = tx.ExecContext(ctx, `DELETE FROM pull_request_reviewers WHERE pull_request_id = $1`, prId)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("delete old reviewers: %w", err)
	}
	for _, r := range newPr.Reviewers {
		_, err := tx.ExecContext(ctx, `INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id) VALUES ($1, $2)`, prId, r.Id)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error inserting new reviewer %s: %w", r.Id, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func (p *PostgresRepo) GetUsersPr(ctx context.Context, userId string, onlyActive bool) ([]entityPr.PullRequest, error) {
	var sb strings.Builder
	sb.WriteString(`
		SELECT 
			pr.pull_request_id,
			pr.pull_request_name,
			pr.author_id,
			pr.status,
			pr.need_more_reviewers,
			pr.created_at,
			pr.merged_at,
			r.reviewer_id
		FROM pull_requests pr
		LEFT JOIN pull_request_reviewers r 
			ON pr.pull_request_id = r.pull_request_id
		WHERE pr.pull_request_id IN (
			SELECT pull_request_id 
			FROM pull_request_reviewers 
			WHERE reviewer_id = $1
		)
	`)
	if onlyActive {
		sb.WriteString(" AND pr.status = 'OPEN'")
	}

	query := sb.String()

	rows, err := p.db.QueryxContext(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("error querying user PRs: %w", err)
	}
	defer rows.Close()

	prMap := make(map[string]*entityPr.PullRequest)

	for rows.Next() {
		var prID, prName, authorID, status, reviewerID sql.NullString
		var needMoreReviewers sql.NullBool
		var createdAt, mergedAt sql.NullTime

		if err := rows.Scan(&prID, &prName, &authorID, &status, &needMoreReviewers, &createdAt, &mergedAt, &reviewerID); err != nil {
			return nil, fmt.Errorf("error scanning PR row: %w", err)
		}

		if _, exists := prMap[prID.String]; !exists {
			prMap[prID.String] = &entityPr.PullRequest{
				Id:                prID.String,
				Name:              prName.String,
				Author:            entityUser.User{Id: authorID.String},
				Status:            status.String,
				NeedMoreReviewers: needMoreReviewers.Bool,
				CreatedAt:         createdAt.Time,
			}
			if mergedAt.Valid {
				prMap[prID.String].MergedAt = &mergedAt.Time
			}
		}
		if reviewerID.Valid {
			prMap[prID.String].Reviewers = append(
				prMap[prID.String].Reviewers,
				entityUser.User{Id: reviewerID.String},
			)
		}
	}

	prs := make([]entityPr.PullRequest, 0, len(prMap))
	for _, pr := range prMap {
		prs = append(prs, *pr)
	}

	return prs, nil
}

func (p *PostgresRepo) GetUserByID(ctx context.Context, userID string) (*entityUser.User, error) {
	var u dto.UserDto
	err := p.db.GetContext(ctx, &u, `
		SELECT user_id, username, team_id, is_active
		FROM users 
		WHERE user_id = $1
	`, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoUserWithId
		}
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	return &entityUser.User{
		Id:       u.Id,
		Name:     u.Name,
		IsActive: u.IsActive,
		TeamID:   u.TeamID,
	}, nil
}

func (p *PostgresRepo) RemoveReviewerFromAllPR(ctx context.Context, reviewerID string) error {
	_, err := p.db.ExecContext(ctx, `
        DELETE FROM pull_request_reviewers
        WHERE reviewer_id = $1
        AND pull_request_id IN (
            SELECT pull_request_id
            FROM pull_requests
            WHERE status != 'MERGED'
        )
    `, reviewerID)
	if err != nil {
		return fmt.Errorf("failed to remove reviewer %s from pull_request_reviewers: %w", reviewerID, err)
	}
	return nil
}

func (p *PostgresRepo) UpdateUser(ctx context.Context, u entityUser.User) error {
	query := `UPDATE users
        SET 
            username = $1,
            team_id = $2,
            is_active = $3
        WHERE user_id = $4`
	res, err := p.db.ExecContext(ctx, query,
		u.Name,
		u.TeamID,
		u.IsActive,
		u.Id,
	)
	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNoUserWithId
	}

	return nil
}

func (p *PostgresRepo) GetUserWithTeam(ctx context.Context, userID string) (*entityUser.User, string, error) {
	query := `SELECT 
            u.user_id,
            u.username,
            u.team_id,
            u.is_active,
            t.team_name
        FROM users u
        LEFT JOIN teams t ON t.id = u.team_id
        WHERE u.user_id = $1`

	var dto struct {
		Id       string `db:"user_id"`
		Name     string `db:"username"`
		TeamID   int    `db:"team_id"`
		IsActive bool   `db:"is_active"`
		TeamName string `db:"team_name"`
	}

	err := p.db.GetContext(ctx, &dto, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", ErrNoUserWithId
		}
		return nil, "", err
	}

	u := &entityUser.User{
		Id:       dto.Id,
		Name:     dto.Name,
		TeamID:   dto.TeamID,
		IsActive: dto.IsActive,
	}

	return u, dto.TeamName, nil
}

func (p *PostgresRepo) AddReviewerToPR(ctx context.Context, prId string, reviewerID string) error {
	_, err := p.db.ExecContext(ctx, `
        INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
        VALUES ($1, $2)
        ON CONFLICT DO NOTHING
    `, prId, reviewerID)
	return err
}

func (p *PostgresRepo) GetTeamPr(ctx context.Context, teamID int) ([]entityPr.PullRequest, error) {
	query := `
		SELECT
			pr.pull_request_id,
			pr.pull_request_name,
			pr.author_id,
			pr.status,
			pr.need_more_reviewers,
			pr.created_at,
			pr.merged_at,
			r.reviewer_id
		FROM pull_requests pr
		LEFT JOIN pull_request_reviewers r
			ON pr.pull_request_id = r.pull_request_id
		WHERE 
			pr.author_id IN (SELECT user_id FROM users WHERE team_id = $1)
			OR
			pr.pull_request_id IN (
				SELECT pull_request_id 
				FROM pull_request_reviewers
				WHERE reviewer_id IN (SELECT user_id FROM users WHERE team_id = $1)
			)
		ORDER BY pr.created_at DESC;
	`

	rows, err := p.db.QueryxContext(ctx, query, teamID)
	if err != nil {
		return nil, fmt.Errorf("query team PRs: %w", err)
	}
	defer rows.Close()

	prMap := make(map[string]*entityPr.PullRequest)

	for rows.Next() {
		var (
			prID, prName, authorID, status sql.NullString
			reviewerID                     sql.NullString
			needMore                       sql.NullBool
			createdAt, mergedAt            sql.NullTime
		)

		if err := rows.Scan(
			&prID, &prName, &authorID, &status,
			&needMore, &createdAt, &mergedAt,
			&reviewerID,
		); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		key := prID.String
		pr, ok := prMap[key]
		if !ok {
			pr = &entityPr.PullRequest{
				Id:                prID.String,
				Name:              prName.String,
				Author:            entityUser.User{Id: authorID.String},
				Status:            status.String,
				NeedMoreReviewers: needMore.Bool,
				CreatedAt:         createdAt.Time,
			}
			if mergedAt.Valid {
				pr.MergedAt = &mergedAt.Time
			}
			prMap[key] = pr
		}

		if reviewerID.Valid {
			pr.Reviewers = append(pr.Reviewers,
				entityUser.User{Id: reviewerID.String},
			)
		}
	}

	result := make([]entityPr.PullRequest, 0, len(prMap))
	for _, pr := range prMap {
		result = append(result, *pr)
	}

	return result, nil
}
