package app

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/anatoly_dev/go-users/internal/domain/auth"
	"github.com/google/uuid"
)

type BruteforceDefenseConfig struct {
	MaxAttempts int

	WindowSeconds int

	BlockDuration int
}

type IPBlockService struct {
	postgresRepo        PostgresIPBlockRepository
	redisRepo           RedisIPBlockRepository
	bruteforceDefConfig BruteforceDefenseConfig
}

type PostgresIPBlockRepository interface {
	Create(ctx context.Context, block *auth.IPBlock) error
	IsBlocked(ctx context.Context, ip net.IP) (bool, *auth.IPBlock, error)
	Remove(ctx context.Context, id uuid.UUID) error
	ListActive(ctx context.Context, offset, limit int) ([]*auth.IPBlock, error)
	GetBlockCount(ctx context.Context, ip net.IP, since time.Time) (int, error)
}

type RedisIPBlockRepository interface {
	Create(ctx context.Context, block *auth.IPBlock) error
	IsBlocked(ctx context.Context, ip net.IP) (bool, *auth.IPBlock, error)
	Remove(ctx context.Context, id uuid.UUID) error
	ListActive(ctx context.Context, offset, limit int) ([]*auth.IPBlock, error)
	GetBlockCount(ctx context.Context, ip net.IP, since time.Time) (int, error)
	RecordLoginAttempt(ctx context.Context, ip net.IP, windowSeconds int) (int, error)
	GetLoginAttempts(ctx context.Context, ip net.IP, windowSeconds int) (int, error)
}

func NewIPBlockService(
	postgresRepo PostgresIPBlockRepository,
	redisRepo RedisIPBlockRepository,
	config BruteforceDefenseConfig,
) *IPBlockService {
	return &IPBlockService{
		postgresRepo:        postgresRepo,
		redisRepo:           redisRepo,
		bruteforceDefConfig: config,
	}
}

func DefaultBruteforceDefenseConfig() BruteforceDefenseConfig {
	return BruteforceDefenseConfig{
		MaxAttempts:   5,
		WindowSeconds: 300,
		BlockDuration: 1800,
	}
}

func (s *IPBlockService) IsBlocked(ctx context.Context, ip net.IP) (bool, *auth.IPBlock, error) {
	isBlocked, block, err := s.postgresRepo.IsBlocked(ctx, ip)
	if err != nil {
		return false, nil, fmt.Errorf("failed to check PostgreSQL IP block: %w", err)
	}

	if isBlocked {
		return true, block, nil
	}

	isBlocked, block, err = s.redisRepo.IsBlocked(ctx, ip)
	if err != nil {
		return false, nil, fmt.Errorf("failed to check Redis IP block: %w", err)
	}

	return isBlocked, block, nil
}

func (s *IPBlockService) CreatePermanentBlock(
	ctx context.Context,
	ip net.IP,
	reason auth.BlockReason,
	createdBy *uuid.UUID,
	comment string,
) (*auth.IPBlock, error) {
	block := auth.NewIPBlock(ip, auth.PermanentBlock, reason, nil, createdBy, comment)

	err := s.postgresRepo.Create(ctx, block)
	if err != nil {
		return nil, fmt.Errorf("failed to create permanent IP block: %w", err)
	}

	return block, nil
}

func (s *IPBlockService) CreateTemporaryBlock(
	ctx context.Context,
	ip net.IP,
	reason auth.BlockReason,
	durationSeconds int,
	createdBy *uuid.UUID,
	comment string,
) (*auth.IPBlock, error) {
	expiresAt := time.Now().Add(time.Duration(durationSeconds) * time.Second)

	block := auth.NewIPBlock(ip, auth.TemporaryBlock, reason, &expiresAt, createdBy, comment)

	err := s.redisRepo.Create(ctx, block)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary IP block: %w", err)
	}

	return block, nil
}

func (s *IPBlockService) RemoveBlock(ctx context.Context, id uuid.UUID) error {
	postgresErr := s.postgresRepo.Remove(ctx, id)

	redisErr := s.redisRepo.Remove(ctx, id)

	if postgresErr == nil || redisErr != nil && redisErr.Error() != fmt.Sprintf("IP block with id %s not found in Redis", id) {
		return postgresErr
	}

	if redisErr == nil || postgresErr != nil && postgresErr.Error() != fmt.Sprintf("IP block with id %s not found", id) {
		return redisErr
	}

	return fmt.Errorf("IP block with id %s not found", id)
}

func (s *IPBlockService) ListActiveBlocks(ctx context.Context, offset, limit int) ([]*auth.IPBlock, error) {
	postgresBlocks, err := s.postgresRepo.ListActive(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list PostgreSQL IP blocks: %w", err)
	}

	redisBlocks, err := s.redisRepo.ListActive(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list Redis IP blocks: %w", err)
	}

	allBlocks := append(postgresBlocks, redisBlocks...)

	if len(allBlocks) > limit {
		return allBlocks[:limit], nil
	}

	return allBlocks, nil
}

func (s *IPBlockService) RecordLoginAttempt(ctx context.Context, ip net.IP) (int, bool, error) {
	isBlocked, _, err := s.IsBlocked(ctx, ip)
	if err != nil {
		return 0, false, fmt.Errorf("failed to check if IP is blocked: %w", err)
	}

	if isBlocked {
		return 0, true, nil
	}

	count, err := s.redisRepo.RecordLoginAttempt(ctx, ip, s.bruteforceDefConfig.WindowSeconds)
	if err != nil {
		return 0, false, fmt.Errorf("failed to record login attempt: %w", err)
	}

	shouldBlock := count >= s.bruteforceDefConfig.MaxAttempts

	if shouldBlock {
		comment := fmt.Sprintf("Automated block after %d failed login attempts within %d seconds",
			count, s.bruteforceDefConfig.WindowSeconds)

		_, err = s.CreateTemporaryBlock(
			ctx,
			ip,
			auth.BruteforceAttempt,
			s.bruteforceDefConfig.BlockDuration,
			nil,
			comment,
		)

		if err != nil {
			return count, false, fmt.Errorf("failed to create temporary block: %w", err)
		}
	}

	return count, shouldBlock, nil
}

func (s *IPBlockService) GetLoginAttempts(ctx context.Context, ip net.IP) (int, error) {
	return s.redisRepo.GetLoginAttempts(ctx, ip, s.bruteforceDefConfig.WindowSeconds)
}

func (s *IPBlockService) ShouldBlockPermanently(ctx context.Context, ip net.IP) (bool, error) {
	since := time.Now().Add(-24 * time.Hour)

	redisCount, err := s.redisRepo.GetBlockCount(ctx, ip, since)
	if err != nil {
		return false, fmt.Errorf("failed to get Redis block count: %w", err)
	}

	postgresCount, err := s.postgresRepo.GetBlockCount(ctx, ip, since)
	if err != nil {
		return false, fmt.Errorf("failed to get PostgreSQL block count: %w", err)
	}

	totalCount := redisCount + postgresCount

	return totalCount >= 3, nil
}
