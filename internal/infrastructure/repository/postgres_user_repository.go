package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/anatoly_dev/go-users/internal/domain/user"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{
		db: db,
	}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *user.User) error {
	query := `
		INSERT INTO users (id, email, hashed_password, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		user.ID,
		user.Email,
		user.HashedPassword,
		user.FirstName,
		user.LastName,
		string(user.Role),
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return fmt.Errorf("user with email %s already exists", user.Email)
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	query := `
		SELECT id, email, hashed_password, first_name, last_name, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)

	return r.scanUser(row)
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `
		SELECT id, email, hashed_password, first_name, last_name, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	row := r.db.QueryRowContext(ctx, query, email)

	return r.scanUser(row)
}

func (r *PostgresUserRepository) Update(ctx context.Context, user *user.User) error {
	query := `
		UPDATE users
		SET email = $2, hashed_password = $3, first_name = $4, last_name = $5, role = $6, updated_at = $7
		WHERE id = $1
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.ID,
		user.Email,
		user.HashedPassword,
		user.FirstName,
		user.LastName,
		string(user.Role),
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (r *PostgresUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM users
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user with id %s not found", id)
	}

	return nil
}

func (r *PostgresUserRepository) List(ctx context.Context, offset, limit int) ([]*user.User, error) {
	query := `
		SELECT id, email, hashed_password, first_name, last_name, role, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*user.User
	for rows.Next() {
		var u user.User
		var role string

		err := rows.Scan(
			&u.ID,
			&u.Email,
			&u.HashedPassword,
			&u.FirstName,
			&u.LastName,
			&role,
			&u.CreatedAt,
			&u.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		u.Role = user.Role(role)
		users = append(users, &u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users rows: %w", err)
	}

	return users, nil
}

func (r *PostgresUserRepository) scanUser(row *sql.Row) (*user.User, error) {
	var u user.User
	var role string

	err := row.Scan(
		&u.ID,
		&u.Email,
		&u.HashedPassword,
		&u.FirstName,
		&u.LastName,
		&role,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}

	u.Role = user.Role(role)
	return &u, nil
}
