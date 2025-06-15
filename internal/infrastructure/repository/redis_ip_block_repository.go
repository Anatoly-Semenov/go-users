package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/anatoly_dev/go-users/internal/domain/auth"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	ipBlockKeyPrefix = "ip_block:"

	loginAttemptsPrefix = "login_attempts:"
)

type RedisIPBlockRepository struct {
	client *redis.Client
}

func NewRedisIPBlockRepository(client *redis.Client) *RedisIPBlockRepository {
	return &RedisIPBlockRepository{
		client: client,
	}
}

func (r *RedisIPBlockRepository) Create(ctx context.Context, block *auth.IPBlock) error {
	if block.ExpiresAt == nil {
		return fmt.Errorf("temporary IP blocks must have expiry time")
	}

	blockData, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("failed to marshal IP block: %w", err)
	}

	expiration := time.Until(*block.ExpiresAt)
	if expiration <= 0 {
		return fmt.Errorf("block expiration must be in the future")
	}

	key := fmt.Sprintf("%s%s", ipBlockKeyPrefix, block.IP.String())
	err = r.client.Set(ctx, key, blockData, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to store IP block in Redis: %w", err)
	}

	return nil
}

func (r *RedisIPBlockRepository) IsBlocked(ctx context.Context, ip net.IP) (bool, *auth.IPBlock, error) {
	key := fmt.Sprintf("%s%s", ipBlockKeyPrefix, ip.String())

	blockData, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil, nil
		}
		return false, nil, fmt.Errorf("failed to check IP block in Redis: %w", err)
	}

	var block auth.IPBlock
	if err = json.Unmarshal([]byte(blockData), &block); err != nil {
		return false, nil, fmt.Errorf("failed to unmarshal IP block data: %w", err)
	}

	return true, &block, nil
}

func (r *RedisIPBlockRepository) Remove(ctx context.Context, id uuid.UUID) error {

	var cursor uint64
	var keys []string

	for {
		var batch []string
		var err error

		batch, cursor, err = r.client.Scan(ctx, cursor, ipBlockKeyPrefix+"*", 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan Redis keys: %w", err)
		}

		keys = append(keys, batch...)

		if cursor == 0 {
			break
		}
	}

	for _, key := range keys {
		blockData, err := r.client.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var block auth.IPBlock
		if err = json.Unmarshal([]byte(blockData), &block); err != nil {
			continue
		}

		if block.ID == id {
			err := r.client.Del(ctx, key).Err()
			if err != nil {
				return fmt.Errorf("failed to delete IP block from Redis: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("IP block with id %s not found in Redis", id)
}

func (r *RedisIPBlockRepository) ListActive(ctx context.Context, offset, limit int) ([]*auth.IPBlock, error) {
	var cursor uint64
	var keys []string

	for {
		var batch []string
		var err error

		batch, cursor, err = r.client.Scan(ctx, cursor, ipBlockKeyPrefix+"*", 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan Redis keys: %w", err)
		}

		keys = append(keys, batch...)

		if cursor == 0 {
			break
		}
	}

	start := offset
	end := offset + limit

	if start >= len(keys) {
		return []*auth.IPBlock{}, nil
	}

	if end > len(keys) {
		end = len(keys)
	}

	pageKeys := keys[start:end]

	blocks := make([]*auth.IPBlock, 0, len(pageKeys))

	for _, key := range pageKeys {
		blockData, err := r.client.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var block auth.IPBlock
		if err = json.Unmarshal([]byte(blockData), &block); err != nil {
			continue
		}

		blocks = append(blocks, &block)
	}

	return blocks, nil
}

func (r *RedisIPBlockRepository) RecordLoginAttempt(ctx context.Context, ip net.IP, windowSeconds int) (int, error) {
	key := fmt.Sprintf("%s%s", loginAttemptsPrefix, ip.String())

	now := time.Now().Unix()
	err := r.client.ZAdd(ctx, key, redis.Z{
		Score:  float64(now),
		Member: now,
	}).Err()

	if err != nil {
		return 0, fmt.Errorf("failed to record login attempt: %w", err)
	}

	r.client.Expire(ctx, key, time.Duration(windowSeconds)*time.Second)

	cutoff := now - int64(windowSeconds)
	r.client.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", cutoff))
	
	count, err := r.client.ZCard(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to count login attempts: %w", err)
	}

	return int(count), nil
}

func (r *RedisIPBlockRepository) GetBlockCount(ctx context.Context, ip net.IP, since time.Time) (int, error) {
	isBlocked, _, err := r.IsBlocked(ctx, ip)
	if err != nil {
		return 0, err
	}

	if isBlocked {
		return 1, nil
	}

	return 0, nil
}

func (r *RedisIPBlockRepository) GetLoginAttempts(ctx context.Context, ip net.IP, windowSeconds int) (int, error) {
	key := fmt.Sprintf("%s%s", loginAttemptsPrefix, ip.String())

	now := time.Now().Unix()
	cutoff := now - int64(windowSeconds)

	count, err := r.client.ZCount(ctx, key, fmt.Sprintf("%d", cutoff), "+inf").Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to count login attempts: %w", err)
	}

	return int(count), nil
}
