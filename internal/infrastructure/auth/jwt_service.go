package auth

import (
	"context"
	"time"

	"github.com/anatoly_dev/go-users/internal/domain/auth"
	"github.com/anatoly_dev/go-users/internal/domain/user"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type JWTService struct {
	userRepo      user.Repository
	secretKey     []byte
	tokenDuration time.Duration
}

func NewJWTService(
	userRepo user.Repository,
	secretKey string,
	tokenDuration time.Duration,
) *JWTService {
	return &JWTService{
		userRepo:      userRepo,
		secretKey:     []byte(secretKey),
		tokenDuration: tokenDuration,
	}
}

func (s *JWTService) HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func (s *JWTService) VerifyPassword(password string, hashedPassword []byte) bool {
	err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	return err == nil
}

func (s *JWTService) GenerateToken(user *user.User, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"role":    string(user.Role),
		"exp":     time.Now().Add(duration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(s.secretKey)
}

func (s *JWTService) ValidateToken(tokenString string) (*auth.TokenPayload, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, auth.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, auth.ErrInvalidToken
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, auth.ErrInvalidToken
	}

	if time.Now().Unix() > int64(exp) {
		return nil, auth.ErrTokenExpired
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, auth.ErrInvalidToken
	}

	email, ok := claims["email"].(string)
	if !ok {
		return nil, auth.ErrInvalidToken
	}

	role, ok := claims["role"].(string)
	if !ok {
		return nil, auth.ErrInvalidToken
	}

	parsedUserID, err := user.ParseID(userID)
	if err != nil {
		return nil, auth.ErrInvalidToken
	}

	return &auth.TokenPayload{
		UserID: parsedUserID,
		Email:  email,
		Role:   user.Role(role),
		Exp:    int64(exp),
	}, nil
}

func (s *JWTService) Authenticate(ctx context.Context, email, password string) (*user.User, string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", auth.ErrInvalidCredentials
	}

	if !s.VerifyPassword(password, user.HashedPassword) {
		return nil, "", auth.ErrInvalidCredentials
	}

	token, err := s.GenerateToken(user, s.tokenDuration)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}
