package http

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/anatoly_dev/go-users/internal/domain/auth"
)

func IPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r)
		ctx := context.WithValue(r.Context(), "client_ip", ip)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractIP(r *http.Request) string {
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		if ip := r.Header.Get(header); ip != "" {
			parts := strings.Split(ip, ",")
			ip = strings.TrimSpace(parts[0])
			return ip
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func IPBlockMiddleware(ipService IPBlockChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extractIP(r)
			parsedIP := net.ParseIP(ip)

			if parsedIP == nil {
				next.ServeHTTP(w, r)
				return
			}

			isBlocked, block, err := ipService.IsBlocked(r.Context(), parsedIP)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			if isBlocked && (block.Type == auth.PermanentBlock || !block.IsExpired()) {
				http.Error(w, "Forbidden: your IP address is blocked", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

type IPBlockChecker interface {
	IsBlocked(ctx context.Context, ip net.IP) (bool, *auth.IPBlock, error)
}
