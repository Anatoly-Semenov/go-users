package auth

import "errors"

var (
	ErrTooManyAttempts    = errors.New("too many login attempts")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrIPBlocked          = errors.New("ip address is blocked")
	ErrTokenExpired       = errors.New("token expired")
	ErrInvalidToken       = errors.New("invalid token")
)
