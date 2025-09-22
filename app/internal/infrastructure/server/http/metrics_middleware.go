package http

import (
	"github.com/anatoly_dev/go-users/app/pkg/metrics"
	"net/http"
	"strings"
	"time"
)

type MetricsMiddleware struct {
	metricsHelper *metrics.MetricsHelper
}

func NewMetricsMiddleware(metricsHelper *metrics.MetricsHelper) *MetricsMiddleware {
	return &MetricsMiddleware{
		metricsHelper: metricsHelper,
	}
}

func (m *MetricsMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		m.metricsHelper.HTTPConcurrentRequests().Inc()
		defer m.metricsHelper.HTTPConcurrentRequests().Dec()

		wrapper := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapper, r)

		duration := time.Since(start)
		endpoint := m.normalizeEndpoint(r.URL.Path)
		statusClass := m.getStatusClass(wrapper.statusCode)

		m.metricsHelper.RecordHTTPRequest(
			r.Method,
			endpoint,
			statusClass,
			duration,
			r.ContentLength,
			int64(wrapper.size),
		)

		switch {
		case strings.Contains(endpoint, "/login"):
			m.metricsHelper.HTTPLoginDuration().Observe(duration.Seconds())
		case strings.Contains(endpoint, "/register"):
			m.metricsHelper.HTTPRegistrationDuration().Observe(duration.Seconds())
		case strings.Contains(endpoint, "/users"):
			m.metricsHelper.HTTPUserLookupDuration().Observe(duration.Seconds())
		}

		if wrapper.statusCode >= 400 {
			errorType := m.categorizeHTTPError(wrapper.statusCode)
			m.metricsHelper.ErrorsTotal().WithLabelValues("http", errorType).Inc()
		}

		if duration.Seconds() > 1.0 {
			m.metricsHelper.SlowResponseTime().WithLabelValues("1s").Set(1)
		} else {
			m.metricsHelper.SlowResponseTime().WithLabelValues("1s").Set(0)
		}

		m.metricsHelper.ResponseTimePercentiles().WithLabelValues("http").Observe(duration.Seconds())
	})
}

func (m *MetricsMiddleware) PanicRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				m.metricsHelper.RecordPanicRecovery()
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func (m *MetricsMiddleware) normalizeEndpoint(path string) string {

	parts := strings.Split(path, "/")
	for i, part := range parts {
		if m.isUUID(part) {
			parts[i] = "{id}"
		}
	}

	normalized := strings.Join(parts, "/")

	switch {
	case strings.HasPrefix(normalized, "/api/users/{id}"):
		return "/api/users/{id}"
	case strings.HasPrefix(normalized, "/api/users"):
		return "/api/users"
	case strings.HasPrefix(normalized, "/api/login"):
		return "/api/login"
	case strings.HasPrefix(normalized, "/api/register"):
		return "/api/register"
	case strings.HasPrefix(normalized, "/metrics"):
		return "/metrics"
	case strings.HasPrefix(normalized, "/health"):
		return "/health"
	default:
		return normalized
	}
}

func (m *MetricsMiddleware) isUUID(s string) bool {
	return len(s) == 36 && strings.Count(s, "-") == 4
}

func (m *MetricsMiddleware) getStatusClass(statusCode int) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return "2xx"
	case statusCode >= 300 && statusCode < 400:
		return "3xx"
	case statusCode >= 400 && statusCode < 500:
		return "4xx"
	case statusCode >= 500:
		return "5xx"
	default:
		return "1xx"
	}
}

func (m *MetricsMiddleware) categorizeHTTPError(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return "bad_request"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusMethodNotAllowed:
		return "method_not_allowed"
	case http.StatusTooManyRequests:
		return "too_many_requests"
	case http.StatusInternalServerError:
		return "internal_server_error"
	case http.StatusBadGateway:
		return "bad_gateway"
	case http.StatusServiceUnavailable:
		return "service_unavailable"
	case http.StatusGatewayTimeout:
		return "gateway_timeout"
	default:
		if statusCode >= 400 && statusCode < 500 {
			return "client_error"
		} else if statusCode >= 500 {
			return "server_error"
		}
		return "unknown"
	}
}
