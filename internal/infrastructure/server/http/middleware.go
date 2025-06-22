package http

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/anatoly_dev/go-users/internal/domain/auth"
	"github.com/anatoly_dev/go-users/pkg/metrics"
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

func IPBlockMiddleware(ipService IPBlockChecker, metricsHelper *metrics.MetricsHelper) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			middlewareName := "ip_block"

			defer func() {
				duration := time.Since(start)
				metricsHelper.RecordMiddlewareProcessing(middlewareName, duration, nil)
			}()

			ip := extractIP(r)
			parsedIP := net.ParseIP(ip)

			if parsedIP == nil {
				next.ServeHTTP(w, r)
				return
			}

			isBlocked, block, err := ipService.IsBlocked(r.Context(), parsedIP)
			if err != nil {
				metricsHelper.RecordMiddlewareProcessing(middlewareName, time.Since(start), err)
				next.ServeHTTP(w, r)
				return
			}

			if isBlocked && (block.Type == auth.PermanentBlock || !block.IsExpired()) {
				blockType := string(block.Type)
				reason := "ip_blocked"

				metricsHelper.RecordBlockedRequest(reason)
				metricsHelper.IPBlocksActive().WithLabelValues(blockType).Inc()

				metricsHelper.RecordSecurityViolation("blocked_ip_access")

				http.Error(w, "Forbidden: your IP address is blocked", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func SecurityMiddleware(metricsHelper *metrics.MetricsHelper) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if containsSuspiciousPatterns(r) {
				metricsHelper.RecordSecurityViolation("suspicious_request")
				metricsHelper.RecordBlockedRequest("suspicious_pattern")
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			if isPotentialDDoS(r) {
				metricsHelper.RecordSecurityViolation("potential_ddos")
			}

			if hasUnusualTrafficPattern(r) {
				metricsHelper.RecordSecurityViolation("unusual_traffic")
			}

			next.ServeHTTP(w, r)
		})
	}
}

func containsSuspiciousPatterns(r *http.Request) bool {
	suspiciousPatterns := []string{
		"<script",
		"javascript:",
		"SELECT * FROM",
		"UNION SELECT",
		"DROP TABLE",
		"../../../",
		"../../../../",
		"cmd.exe",
		"/bin/sh",
		"eval(",
		"base64_decode",
	}
	
	path := strings.ToLower(r.URL.Path)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(path, strings.ToLower(pattern)) {
			return true
		}
	}

	query := strings.ToLower(r.URL.RawQuery)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(query, strings.ToLower(pattern)) {
			return true
		}
	}

	for name, values := range r.Header {
		headerName := strings.ToLower(name)
		if headerName == "user-agent" || headerName == "referer" {
			continue
		}

		for _, value := range values {
			lowerValue := strings.ToLower(value)
			for _, pattern := range suspiciousPatterns {
				if strings.Contains(lowerValue, strings.ToLower(pattern)) {
					return true
				}
			}
		}
	}

	return false
}

func isPotentialDDoS(r *http.Request) bool {

	userAgent := r.Header.Get("User-Agent")
	if userAgent == "" || len(userAgent) < 10 {
		return true
	}

	if r.Header.Get("Accept") == "" && r.Header.Get("Accept-Language") == "" {
		return true
	}

	return false
}

func hasUnusualTrafficPattern(r *http.Request) bool {
	path := r.URL.Path

	if strings.Contains(path, "//") {
		return true
	}

	if len(path) > 1000 {
		return true
	}

	if len(r.URL.Query()) > 50 {
		return true
	}

	return false
}

type IPBlockChecker interface {
	IsBlocked(ctx context.Context, ip net.IP) (bool, *auth.IPBlock, error)
}
