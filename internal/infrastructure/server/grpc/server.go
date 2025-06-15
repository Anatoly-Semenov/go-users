package grpc

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/anatoly_dev/go-users/internal/app"
	"github.com/anatoly_dev/go-users/internal/config"
	"github.com/anatoly_dev/go-users/internal/domain/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"

	userv1 "github.com/anatoly_dev/go-users/api/v1"
)

type Server struct {
	userService *app.UserService
	config      *config.Config
	server      *grpc.Server
	userv1.UnimplementedUserServiceServer
}

func NewServer(userService *app.UserService, config *config.Config) *Server {
	return &Server{
		userService: userService,
		config:      config,
	}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%s", s.config.Server.GRPCPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.server = grpc.NewServer()
	userv1.RegisterUserServiceServer(s.server, s)

	reflection.Register(s.server)

	log.Printf("Starting gRPC server on port %s", s.config.Server.GRPCPort)
	return s.server.Serve(listener)
}

func (s *Server) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
		log.Printf("Stopped gRPC server")
	}
}

func (s *Server) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	role := user.RoleUser
	if req.Role == userv1.UserRole_USER_ROLE_ADMIN {
		role = user.RoleAdmin
	}

	params := app.RegisterUserParams{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      role,
	}

	createdUser, err := s.userService.Register(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	return &userv1.RegisterResponse{
		User: mapDomainUserToProto(createdUser),
	}, nil
}

func (s *Server) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	user, token, err := s.userService.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to login: %w", err)
	}

	return &userv1.LoginResponse{
		User:  mapDomainUserToProto(user),
		Token: token,
	}, nil
}

func (s *Server) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	id, err := user.ParseID(req.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	foundUser, err := s.userService.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &userv1.GetUserResponse{
		User: mapDomainUserToProto(foundUser),
	}, nil
}

func (s *Server) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	id, err := user.ParseID(req.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	params := app.UpdateUserParams{
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	updatedUser, err := s.userService.Update(ctx, id, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &userv1.UpdateUserResponse{
		User: mapDomainUserToProto(updatedUser),
	}, nil
}

func (s *Server) DeleteUser(ctx context.Context, req *userv1.DeleteUserRequest) (*userv1.DeleteUserResponse, error) {
	id, err := user.ParseID(req.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	if err := s.userService.Delete(ctx, id); err != nil {
		return nil, fmt.Errorf("failed to delete user: %w", err)
	}

	return &userv1.DeleteUserResponse{}, nil
}

func (s *Server) ListUsers(ctx context.Context, req *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	users, err := s.userService.List(ctx, int(req.Offset), int(req.Limit))
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	var protoUsers []*userv1.User
	for _, u := range users {
		protoUsers = append(protoUsers, mapDomainUserToProto(u))
	}

	return &userv1.ListUsersResponse{
		Users: protoUsers,
	}, nil
}

func mapDomainUserToProto(u *user.User) *userv1.User {
	var role userv1.UserRole
	switch u.Role {
	case user.RoleAdmin:
		role = userv1.UserRole_USER_ROLE_ADMIN
	default:
		role = userv1.UserRole_USER_ROLE_USER
	}

	return &userv1.User{
		Id:        u.ID.String(),
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Role:      role,
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}
}
