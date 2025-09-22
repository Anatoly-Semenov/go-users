package repository

import (
	"context"
	"github.com/anatoly_dev/go-users/app/internal/domain/auth"
	"github.com/anatoly_dev/go-users/app/pkg/metrics"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisIPBlockRepositoryWithMetrics struct {
	repo          *RedisIPBlockRepository
	metricsHelper *metrics.MetricsHelper
}

func NewRedisIPBlockRepositoryWithMetrics(client *redis.Client, metricsHelper *metrics.MetricsHelper) *RedisIPBlockRepositoryWithMetrics {
	return &RedisIPBlockRepositoryWithMetrics{
		repo:          NewRedisIPBlockRepository(client),
		metricsHelper: metricsHelper,
	}
}

func (r *RedisIPBlockRepositoryWithMetrics) GetBlockCount(ctx context.Context, ip net.IP, since time.Time) (int, error) {
	start := time.Now()
	defer func() {
		r.metricsHelper.RecordRedisOperation("get", time.Since(start), nil)
	}()

	count, err := r.repo.GetBlockCount(ctx, ip, since)
	if err != nil {
		r.metricsHelper.RecordRedisOperation("get", time.Since(start), err)
	}
	return count, err
}

func (r *RedisIPBlockRepositoryWithMetrics) ListActive(ctx context.Context, offset, limit int) ([]*auth.IPBlock, error) {
	start := time.Now()
	defer func() {
		r.metricsHelper.RecordRedisOperation("scan", time.Since(start), nil)
	}()

	blocks, err := r.repo.ListActive(ctx, offset, limit)
	if err != nil {
		r.metricsHelper.RecordRedisOperation("scan", time.Since(start), err)
	}
	return blocks, err
}

func (r *RedisIPBlockRepositoryWithMetrics) IsBlocked(ctx context.Context, ip net.IP) (bool, *auth.IPBlock, error) {
	start := time.Now()
	defer func() {
		r.metricsHelper.RecordRedisOperation("get", time.Since(start), nil)
	}()

	isBlocked, block, err := r.repo.IsBlocked(ctx, ip)

	if err == nil && isBlocked {
		r.metricsHelper.RedisIPBlockHitsTotal().Inc()
	} else if err == nil && !isBlocked {
		r.metricsHelper.RedisIPBlockMissesTotal().Inc()
	}

	if err != nil {
		r.metricsHelper.RecordRedisOperation("get", time.Since(start), err)
	}

	return isBlocked, block, err
}

func (r *RedisIPBlockRepositoryWithMetrics) Create(ctx context.Context, block *auth.IPBlock) error {
	start := time.Now()
	defer func() {
		r.metricsHelper.RecordRedisOperation("set", time.Since(start), nil)
		r.metricsHelper.RedisTemporaryBlocksCreated().Inc()
	}()

	err := r.repo.Create(ctx, block)
	if err != nil {
		r.metricsHelper.RecordRedisOperation("set", time.Since(start), err)
	}
	return err
}

func (r *RedisIPBlockRepositoryWithMetrics) Remove(ctx context.Context, id uuid.UUID) error {
	start := time.Now()
	defer func() {
		r.metricsHelper.RecordRedisOperation("del", time.Since(start), nil)
	}()

	err := r.repo.Remove(ctx, id)
	if err != nil {
		r.metricsHelper.RecordRedisOperation("del", time.Since(start), err)
	}
	return err
}

func (r *RedisIPBlockRepositoryWithMetrics) RecordLoginAttempt(ctx context.Context, ip net.IP, windowSeconds int) (int, error) {
	start := time.Now()
	defer func() {
		r.metricsHelper.RecordRedisOperation("zadd", time.Since(start), nil)
	}()

	count, err := r.repo.RecordLoginAttempt(ctx, ip, windowSeconds)
	if err != nil {
		r.metricsHelper.RecordRedisOperation("zadd", time.Since(start), err)
	}
	return count, err
}

func (r *RedisIPBlockRepositoryWithMetrics) GetLoginAttempts(ctx context.Context, ip net.IP, windowSeconds int) (int, error) {
	start := time.Now()
	defer func() {
		r.metricsHelper.RecordRedisOperation("zcount", time.Since(start), nil)
	}()

	count, err := r.repo.GetLoginAttempts(ctx, ip, windowSeconds)
	if err != nil {
		r.metricsHelper.RecordRedisOperation("zcount", time.Since(start), err)
	}
	return count, err
}

func (r *RedisIPBlockRepositoryWithMetrics) UpdateMetrics(ctx context.Context) {
	r.metricsHelper.UpdateRedisMetrics(ctx)
}
