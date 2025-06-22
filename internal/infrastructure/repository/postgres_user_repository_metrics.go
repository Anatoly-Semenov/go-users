package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/anatoly_dev/go-users/internal/domain/user"
	"github.com/anatoly_dev/go-users/pkg/metrics"
	"github.com/google/uuid"
)

type PostgresUserRepositoryWithMetrics struct {
	repo          *PostgresUserRepository
	metricsHelper *metrics.MetricsHelper
}

func NewPostgresUserRepositoryWithMetrics(db *sql.DB, metricsHelper *metrics.MetricsHelper) *PostgresUserRepositoryWithMetrics {
	return &PostgresUserRepositoryWithMetrics{
		repo:          NewPostgresUserRepository(db),
		metricsHelper: metricsHelper,
	}
}

func (r *PostgresUserRepositoryWithMetrics) Create(ctx context.Context, user *user.User) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		r.metricsHelper.RecordDBOperation("insert", "users", duration, 1, nil)
	}()

	err := r.repo.Create(ctx, user)
	if err != nil {
		r.metricsHelper.RecordDBOperation("insert", "users", time.Since(start), 0, err)
	}
	return err
}

func (r *PostgresUserRepositoryWithMetrics) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	start := time.Now()
	defer func() {
		r.metricsHelper.DBUserQueriesTotal().WithLabelValues("by_id").Inc()
	}()

	user, err := r.repo.GetByID(ctx, id)
	r.metricsHelper.RecordDBOperation("select", "users", time.Since(start), 1, err)
	return user, err
}

func (r *PostgresUserRepositoryWithMetrics) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	start := time.Now()
	defer func() {
		r.metricsHelper.DBUserQueriesTotal().WithLabelValues("by_email").Inc()
	}()

	user, err := r.repo.GetByEmail(ctx, email)
	r.metricsHelper.RecordDBOperation("select", "users", time.Since(start), 1, err)
	return user, err
}

func (r *PostgresUserRepositoryWithMetrics) Update(ctx context.Context, user *user.User) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		r.metricsHelper.RecordDBOperation("update", "users", duration, 1, nil)
	}()

	err := r.repo.Update(ctx, user)
	if err != nil {
		r.metricsHelper.RecordDBOperation("update", "users", time.Since(start), 0, err)
	}
	return err
}

func (r *PostgresUserRepositoryWithMetrics) Delete(ctx context.Context, id uuid.UUID) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		r.metricsHelper.RecordDBOperation("delete", "users", duration, 1, nil)
	}()

	err := r.repo.Delete(ctx, id)
	if err != nil {
		r.metricsHelper.RecordDBOperation("delete", "users", time.Since(start), 0, err)
	}
	return err
}

func (r *PostgresUserRepositoryWithMetrics) List(ctx context.Context, offset, limit int) ([]*user.User, error) {
	start := time.Now()
	defer func() {
		r.metricsHelper.DBUserQueriesTotal().WithLabelValues("list").Inc()
	}()

	users, err := r.repo.List(ctx, offset, limit)
	rowsAffected := int64(len(users))
	if err != nil {
		rowsAffected = 0
	}
	r.metricsHelper.RecordDBOperation("select", "users", time.Since(start), rowsAffected, err)
	return users, err
}
