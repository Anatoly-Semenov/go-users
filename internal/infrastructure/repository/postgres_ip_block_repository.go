package repository

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"time"

	"github.com/anatoly_dev/go-users/internal/domain/auth"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type PostgresIPBlockRepository struct {
	db *sql.DB
}

func NewPostgresIPBlockRepository(db *sql.DB) *PostgresIPBlockRepository {
	return &PostgresIPBlockRepository{
		db: db,
	}
}

func (r *PostgresIPBlockRepository) Create(ctx context.Context, block *auth.IPBlock) error {
	query := `
		INSERT INTO ip_blocks (id, ip, block_type, reason, created_at, expires_at, created_by, comment)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		block.ID,
		block.IP.String(),
		string(block.Type),
		string(block.Reason),
		block.CreatedAt,
		block.ExpiresAt,
		block.CreatedBy,
		block.Comment,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return fmt.Errorf("IP %s is already blocked", block.IP.String())
		}
		return fmt.Errorf("failed to create IP block: %w", err)
	}

	return nil
}

func (r *PostgresIPBlockRepository) IsBlocked(ctx context.Context, ip net.IP) (bool, *auth.IPBlock, error) {
	query := `
		SELECT id, ip, block_type, reason, created_at, expires_at, created_by, comment 
		FROM ip_blocks 
		WHERE ip = $1 AND (expires_at IS NULL OR expires_at > NOW())
	`

	row := r.db.QueryRowContext(ctx, query, ip.String())

	block, err := r.scanIPBlock(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil, nil
		}
		return false, nil, fmt.Errorf("failed to check IP block status: %w", err)
	}

	return true, block, nil
}

func (r *PostgresIPBlockRepository) Remove(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM ip_blocks WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete IP block: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("IP block with id %s not found", id)
	}

	return nil
}

func (r *PostgresIPBlockRepository) ListActive(ctx context.Context, offset, limit int) ([]*auth.IPBlock, error) {
	query := `
		SELECT id, ip, block_type, reason, created_at, expires_at, created_by, comment 
		FROM ip_blocks 
		WHERE expires_at IS NULL OR expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query IP blocks: %w", err)
	}
	defer rows.Close()

	var blocks []*auth.IPBlock
	for rows.Next() {
		block, err := r.scanIPBlockFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan IP block: %w", err)
		}
		blocks = append(blocks, block)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating IP blocks rows: %w", err)
	}

	return blocks, nil
}

func (r *PostgresIPBlockRepository) GetBlockCount(ctx context.Context, ip net.IP, since time.Time) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM ip_blocks 
		WHERE ip = $1 AND created_at >= $2
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, ip.String(), since).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get IP block count: %w", err)
	}

	return count, nil
}

func (r *PostgresIPBlockRepository) scanIPBlock(row *sql.Row) (*auth.IPBlock, error) {
	var block auth.IPBlock
	var ipStr string
	var blockType, reason string

	err := row.Scan(
		&block.ID,
		&ipStr,
		&blockType,
		&reason,
		&block.CreatedAt,
		&block.ExpiresAt,
		&block.CreatedBy,
		&block.Comment,
	)
	if err != nil {
		return nil, err
	}

	block.IP = net.ParseIP(ipStr)
	block.Type = auth.BlockType(blockType)
	block.Reason = auth.BlockReason(reason)

	return &block, nil
}

func (r *PostgresIPBlockRepository) scanIPBlockFromRows(rows *sql.Rows) (*auth.IPBlock, error) {
	var block auth.IPBlock
	var ipStr string
	var blockType, reason string

	err := rows.Scan(
		&block.ID,
		&ipStr,
		&blockType,
		&reason,
		&block.CreatedAt,
		&block.ExpiresAt,
		&block.CreatedBy,
		&block.Comment,
	)
	if err != nil {
		return nil, err
	}

	block.IP = net.ParseIP(ipStr)
	block.Type = auth.BlockType(blockType)
	block.Reason = auth.BlockReason(reason)

	return &block, nil
}
