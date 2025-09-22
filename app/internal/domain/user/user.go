package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type User struct {
	ID             uuid.UUID
	Email          string
	HashedPassword []byte
	FirstName      string
	LastName       string
	Role           Role
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewUser(email, firstName, lastName string, role Role) (*User, error) {
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}

	return &User{
		ID:        uuid.New(),
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (u *User) SetPassword(hashedPassword []byte) {
	u.HashedPassword = hashedPassword
	u.UpdatedAt = time.Now()
}

func (u *User) Update(firstName, lastName string) {
	if firstName != "" {
		u.FirstName = firstName
	}

	if lastName != "" {
		u.LastName = lastName
	}

	u.UpdatedAt = time.Now()
}

func ParseID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}
