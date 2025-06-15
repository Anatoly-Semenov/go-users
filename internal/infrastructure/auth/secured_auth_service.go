package auth

import (
	"context"
	"net"
	"time"

	"github.com/anatoly_dev/go-users/internal/app"
	"github.com/anatoly_dev/go-users/internal/domain/auth"
	"github.com/anatoly_dev/go-users/internal/domain/user"
)

type SecuredAuthService struct {
	baseService    auth.Service
	ipBlockService *app.IPBlockService
}

func NewSecuredAuthService(
	baseService auth.Service,
	ipBlockService *app.IPBlockService,
) *SecuredAuthService {
	return &SecuredAuthService{
		baseService:    baseService,
		ipBlockService: ipBlockService,
	}
}

func GetClientIP(ctx context.Context) net.IP {
	ipStr, ok := ctx.Value("client_ip").(string)
	if !ok || ipStr == "" {
		return net.IPv4(127, 0, 0, 1)
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return net.IPv4(127, 0, 0, 1)
	}

	return ip
}

func (s *SecuredAuthService) Authenticate(ctx context.Context, email, password string) (*user.User, string, error) {
	clientIP := GetClientIP(ctx)

	isBlocked, block, err := s.ipBlockService.IsBlocked(ctx, clientIP)
	if err != nil {
		return nil, "", err
	}

	if isBlocked {
		if block.Type == auth.PermanentBlock || !block.IsExpired() {
			return nil, "", auth.ErrIPBlocked
		}
	}

	_, shouldBlock, err := s.ipBlockService.RecordLoginAttempt(ctx, clientIP)
	if err != nil {
		return nil, "", err
	}

	if shouldBlock {
		return nil, "", auth.ErrTooManyAttempts
	}

	user, token, err := s.baseService.Authenticate(ctx, email, password)

	if err == auth.ErrInvalidCredentials {
		shouldBlockPermanently, checkErr := s.ipBlockService.ShouldBlockPermanently(ctx, clientIP)
		if checkErr == nil && shouldBlockPermanently {
			comment := "Automatic permanent block after multiple temporary blocks"
			_, _ = s.ipBlockService.CreatePermanentBlock(ctx, clientIP, auth.BruteforceAttempt, nil, comment)
		}

		return nil, "", err
	}

	return user, token, err
}

func (s *SecuredAuthService) GenerateToken(user *user.User, duration time.Duration) (string, error) {
	return s.baseService.GenerateToken(user, duration)
}

func (s *SecuredAuthService) VerifyPassword(password string, hashedPassword []byte) bool {
	return s.baseService.VerifyPassword(password, hashedPassword)
}

func (s *SecuredAuthService) ValidateToken(token string) (*auth.TokenPayload, error) {
	return s.baseService.ValidateToken(token)
}

func (s *SecuredAuthService) HashPassword(password string) ([]byte, error) {
	return s.baseService.HashPassword(password)
}
