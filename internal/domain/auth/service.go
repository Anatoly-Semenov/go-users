package auth

import (
	"context"
	"time"

	"github.com/anatoly_dev/go-users/internal/domain/user"
	"github.com/google/uuid"
)

type TokenPayload struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   user.Role `json:"role"`
	Exp    int64     `json:"exp"`
}

type Service interface {
	Authenticate(ctx context.Context, email, password string) (*user.User, string, error)
	GenerateToken(user *user.User, duration time.Duration) (string, error)
	VerifyPassword(password string, hashedPassword []byte) bool
	ValidateToken(token string) (*TokenPayload, error)
	HashPassword(password string) ([]byte, error)
}
