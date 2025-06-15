package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/anatoly_dev/go-users/internal/domain/auth"
	"github.com/anatoly_dev/go-users/internal/domain/user"
	"github.com/google/uuid"
)

type RegisterUserParams struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	Role      user.Role
}

type UpdateUserParams struct {
	FirstName string
	LastName  string
}

type UserService struct {
	userRepo    user.Repository
	authService auth.Service
}

func NewUserService(userRepo user.Repository, authService auth.Service) *UserService {
	return &UserService{
		userRepo:    userRepo,
		authService: authService,
	}
}

func (s *UserService) Register(ctx context.Context, params RegisterUserParams) (*user.User, error) {
	existingUser, err := s.userRepo.GetByEmail(ctx, params.Email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", params.Email)
	}

	newUser, err := user.NewUser(
		params.Email,
		params.FirstName,
		params.LastName,
		params.Role,
	)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := s.authService.HashPassword(params.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	newUser.SetPassword(hashedPassword)

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return newUser, nil
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	return s.userRepo.GetByEmail(ctx, email)
}

func (s *UserService) Update(ctx context.Context, id uuid.UUID, params UpdateUserParams) (*user.User, error) {
	foundUser, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	foundUser.Update(params.FirstName, params.LastName)

	if err := s.userRepo.Update(ctx, foundUser); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return foundUser, nil
}

func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.userRepo.Delete(ctx, id)
}

func (s *UserService) List(ctx context.Context, offset, limit int) ([]*user.User, error) {
	return s.userRepo.List(ctx, offset, limit)
}

func (s *UserService) Login(ctx context.Context, email, password string) (*user.User, string, error) {
	if email == "" || password == "" {
		return nil, "", errors.New("email and password are required")
	}

	return s.authService.Authenticate(ctx, email, password)
}
