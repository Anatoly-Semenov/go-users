package auth

import (
	"context"
	auth2 "github.com/anatoly_dev/go-users/app/internal/domain/auth"
	user2 "github.com/anatoly_dev/go-users/app/internal/domain/user"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type JWTService struct {
	userRepo      user2.Repository
	secretKey     []byte
	tokenDuration time.Duration
}

func NewJWTService(
	userRepo user2.Repository,
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

func (s *JWTService) GenerateToken(user *user2.User, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"role":    string(user.Role),
		"exp":     time.Now().Add(duration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(s.secretKey)
}

func (s *JWTService) ValidateToken(tokenString string) (*auth2.TokenPayload, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, auth2.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, auth2.ErrInvalidToken
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, auth2.ErrInvalidToken
	}

	if time.Now().Unix() > int64(exp) {
		return nil, auth2.ErrTokenExpired
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, auth2.ErrInvalidToken
	}

	email, ok := claims["email"].(string)
	if !ok {
		return nil, auth2.ErrInvalidToken
	}

	role, ok := claims["role"].(string)
	if !ok {
		return nil, auth2.ErrInvalidToken
	}

	parsedUserID, err := user2.ParseID(userID)
	if err != nil {
		return nil, auth2.ErrInvalidToken
	}

	return &auth2.TokenPayload{
		UserID: parsedUserID,
		Email:  email,
		Role:   user2.Role(role),
		Exp:    int64(exp),
	}, nil
}

func (s *JWTService) Authenticate(ctx context.Context, email, password string) (*user2.User, string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", auth2.ErrInvalidCredentials
	}

	if !s.VerifyPassword(password, user.HashedPassword) {
		return nil, "", auth2.ErrInvalidCredentials
	}

	token, err := s.GenerateToken(user, s.tokenDuration)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}
