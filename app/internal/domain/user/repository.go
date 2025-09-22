package user

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	List(ctx context.Context, offset, limit int) ([]*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, user *User) error
	Create(ctx context.Context, user *User) error
}
