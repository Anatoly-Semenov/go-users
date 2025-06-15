package auth

import (
	"net"
	"time"

	"github.com/google/uuid"
)

type BlockType string

const (
	PermanentBlock BlockType = "permanent"

	TemporaryBlock BlockType = "temporary"
)

type BlockReason string

const (
	SuspiciousActivity BlockReason = "suspicious_activity"
	BruteforceAttempt  BlockReason = "bruteforce_attempt"
	ManualBlock        BlockReason = "manual"
)

type IPBlock struct {
	ID        uuid.UUID
	IP        net.IP
	Type      BlockType
	Reason    BlockReason
	CreatedAt time.Time
	ExpiresAt *time.Time
	CreatedBy *uuid.UUID
	Comment   string
}

func NewIPBlock(ip net.IP, blockType BlockType, reason BlockReason,
	expiration *time.Time, createdBy *uuid.UUID, comment string) *IPBlock {

	return &IPBlock{
		ID:        uuid.New(),
		IP:        ip,
		Type:      blockType,
		Reason:    reason,
		CreatedAt: time.Now(),
		ExpiresAt: expiration,
		CreatedBy: createdBy,
		Comment:   comment,
	}
}

func (b *IPBlock) IsExpired() bool {
	if b.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*b.ExpiresAt)
}

type IPBlockRepository interface {
	GetBlockCount(ip net.IP, since time.Time) (int, error)
	ListActive(offset, limit int) ([]*IPBlock, error)
	IsBlocked(ip net.IP) (bool, *IPBlock, error)
	Create(block *IPBlock) error
	Remove(id uuid.UUID) error
}
