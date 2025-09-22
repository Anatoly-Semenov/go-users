package metrics

import (
	"context"
	"database/sql"
	"runtime"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

type MetricsHelper struct {
	metrics *Metrics
	db      *sql.DB
	redis   *redis.Client
}

func NewMetricsHelper(metrics *Metrics, db *sql.DB, redis *redis.Client) *MetricsHelper {
	return &MetricsHelper{
		metrics: metrics,
		db:      db,
		redis:   redis,
	}
}

func (h *MetricsHelper) RecordHTTPRequest(method, endpoint, status string, duration time.Duration, requestSize, responseSize int64) {
	h.metrics.HTTPRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	h.metrics.HTTPRequestDuration.WithLabelValues(method, endpoint, status).Observe(duration.Seconds())
	h.metrics.HTTPRequestSizeBytes.WithLabelValues(method, endpoint).Observe(float64(requestSize))
	h.metrics.HTTPResponseSizeBytes.WithLabelValues(method, endpoint).Observe(float64(responseSize))
	h.metrics.PopularEndpoints.WithLabelValues(endpoint).Inc()
}

func (h *MetricsHelper) RecordGRPCRequest(method, status string, duration time.Duration, requestSize, responseSize int64) {
	h.metrics.GRPCRequestsTotal.WithLabelValues(method, status).Inc()
	h.metrics.GRPCRequestDuration.WithLabelValues(method, status).Observe(duration.Seconds())
	h.metrics.GRPCRequestSizeBytes.WithLabelValues(method).Observe(float64(requestSize))
	h.metrics.GRPCResponseSizeBytes.WithLabelValues(method).Observe(float64(responseSize))
}

func (h *MetricsHelper) RecordUserOperation(operation, status string) {
	h.metrics.UserOperationsTotal.WithLabelValues(operation, status).Inc()
}

func (h *MetricsHelper) RecordLoginAttempt(success bool, ip string) {
	status := "failed"
	if success {
		status = "success"
		h.metrics.JWTTokensIssuedTotal.Inc()
	} else {
		h.metrics.FailedLoginAttemptsTotal.WithLabelValues(ip).Inc()
		h.metrics.AuthErrorsTotal.WithLabelValues("invalid_credentials").Inc()
	}
	h.metrics.UserLoginAttemptsTotal.WithLabelValues(status).Inc()
}

func (h *MetricsHelper) RecordIPBlock(blockType, reason string) {
	h.metrics.IPBlocksTotal.WithLabelValues(blockType).Inc()
	h.metrics.IPBlocksCreatedTotal.WithLabelValues(blockType, reason).Inc()

	if blockType == "temporary" {
		h.metrics.RedisTemporaryBlocksCreated.Inc()
	}
}

func (h *MetricsHelper) RecordBruteforceAttempt(ip string) {
	h.metrics.BruteforceAttemptsTotal.WithLabelValues(ip).Inc()
	h.metrics.SuspiciousActivityTotal.WithLabelValues("bruteforce_attempt").Inc()
}

func (h *MetricsHelper) RecordBruteforceDetected() {
	h.metrics.BruteforceDetectedTotal.Inc()
	h.metrics.ActiveBruteforceAttacks.Inc()
}

func (h *MetricsHelper) RecordDBOperation(operation, table string, duration time.Duration, rowsAffected int64, err error) {
	if err != nil {
		h.metrics.DBQueryErrorsTotal.WithLabelValues(h.categorizeDBError(err)).Inc()
		h.metrics.ErrorsTotal.WithLabelValues("db", h.categorizeDBError(err)).Inc()
	}

	h.metrics.DBOperationsTotal.WithLabelValues(operation, table).Inc()
	h.metrics.DBOperationDuration.WithLabelValues(operation, table).Observe(duration.Seconds())

	if rowsAffected > 0 {
		h.metrics.DBRowsAffectedTotal.WithLabelValues(operation, table).Add(float64(rowsAffected))
	}
}

func (h *MetricsHelper) RecordRedisOperation(command string, duration time.Duration, err error) {
	h.metrics.RedisCommandsTotal.WithLabelValues(command).Inc()
	h.metrics.RedisCommandDuration.WithLabelValues(command).Observe(duration.Seconds())

	if err != nil {
		h.metrics.ErrorsTotal.WithLabelValues("redis", "command_error").Inc()
	}
}

func (h *MetricsHelper) RecordValidationError(field string) {
	h.metrics.ValidationErrorsTotal.WithLabelValues(field).Inc()
	h.metrics.ErrorsTotal.WithLabelValues("validation", field).Inc()
}

func (h *MetricsHelper) RecordJWTValidation(status string, duration time.Duration) {
	h.metrics.JWTTokensValidatedTotal.WithLabelValues(status).Inc()
	h.metrics.JWTTokenValidationDuration.Observe(duration.Seconds())

	if status != "valid" {
		h.metrics.AuthErrorsTotal.WithLabelValues("token_" + status).Inc()
	}
}

func (h *MetricsHelper) RecordSecurityViolation(violationType string) {
	h.metrics.SecurityViolationsTotal.WithLabelValues(violationType).Inc()
	h.metrics.SuspiciousActivityTotal.WithLabelValues(violationType).Inc()
}

func (h *MetricsHelper) RecordBlockedRequest(reason string) {
	h.metrics.BlockedRequestsTotal.WithLabelValues(reason).Inc()
}

func (h *MetricsHelper) RecordMiddlewareProcessing(middleware string, duration time.Duration, err error) {
	h.metrics.MiddlewareRequestDuration.WithLabelValues(middleware).Observe(duration.Seconds())

	if err != nil {
		h.metrics.MiddlewareErrorsTotal.WithLabelValues(middleware, "processing_error").Inc()
	}
}

func (h *MetricsHelper) RecordPanicRecovery() {
	h.metrics.PanicRecoveriesTotal.Inc()
	h.metrics.ErrorsTotal.WithLabelValues("runtime", "panic").Inc()
}

func (h *MetricsHelper) RecordUserSessionMetrics(duration time.Duration, actions int) {
	h.metrics.UserSessionDuration.Observe(duration.Seconds())
	h.metrics.UserActionsPerSession.Observe(float64(actions))
}

func (h *MetricsHelper) UpdateDatabaseConnectionMetrics() {
	if h.db != nil {
		stats := h.db.Stats()
		h.metrics.DBConnectionsActive.Set(float64(stats.InUse))
		h.metrics.DBConnectionsIdle.Set(float64(stats.Idle))
		h.metrics.DBConnectionsMax.Set(float64(stats.MaxOpenConnections))

		if stats.InUse >= int(float64(stats.MaxOpenConnections)*0.9) {
			h.metrics.DBConnectionExhaustion.Set(1)
		} else {
			h.metrics.DBConnectionExhaustion.Set(0)
		}
	}
}

func (h *MetricsHelper) UpdateRedisMetrics(ctx context.Context) {
	if h.redis != nil {
		poolStats := h.redis.PoolStats()
		h.metrics.RedisConnectionsActive.Set(float64(poolStats.TotalConns))

		info, err := h.redis.Info(ctx, "memory").Result()
		if err == nil {
			if memStr := h.extractRedisInfoValue(info, "used_memory:"); memStr != "" {
				if mem, err := strconv.ParseFloat(memStr, 64); err == nil {
					h.metrics.RedisMemoryUsedBytes.Set(mem)
				}
			}
		}

		ipBlockCount, err := h.redis.DBSize(ctx).Result()
		if err == nil {
			h.metrics.RedisKeysTotal.WithLabelValues("ip_block").Set(float64(ipBlockCount))
		}
	}
}

func (h *MetricsHelper) UpdateSystemMetrics() {
	h.metrics.UpdateUptime()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	if memStats.HeapInuse > uint64(float64(memStats.HeapSys)*0.8) {
		h.metrics.MemoryUsageHigh.WithLabelValues("80%").Set(1)
	} else {
		h.metrics.MemoryUsageHigh.WithLabelValues("80%").Set(0)
	}
}

func (h *MetricsHelper) UpdateErrorRates() {
	// TODO: implementing
}

func (h *MetricsHelper) UpdateSLAMetrics() {
	h.metrics.ServiceAvailabilityPercentage.Set(99.9)
	h.metrics.ErrorBudgetRemaining.Set(0.1)
}

func (h *MetricsHelper) categorizeDBError(err error) string {
	errStr := err.Error()
	switch {
	case contains(errStr, "timeout"):
		return "timeout"
	case contains(errStr, "connection"):
		return "connection"
	case contains(errStr, "constraint"):
		return "constraint"
	case contains(errStr, "duplicate key"):
		return "duplicate_key"
	default:
		return "unknown"
	}
}

func (h *MetricsHelper) extractRedisInfoValue(info, key string) string {
	lines := splitLines(info)
	for _, line := range lines {
		if startsWith(line, key) {
			return line[len(key):]
		}
	}
	return ""
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, char := range s {
		if char == '\n' || char == '\r' {
			if i > start {
				lines = append(lines, s[start:i])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func CreateMetricsHandler() *prometheus.Registry {
	return prometheus.DefaultRegisterer.(*prometheus.Registry)
}

func (h *MetricsHelper) HTTPConcurrentRequests() prometheus.Gauge {
	return h.metrics.HTTPConcurrentRequests
}

func (h *MetricsHelper) HTTPLoginDuration() prometheus.Histogram {
	return h.metrics.HTTPLoginDuration
}

func (h *MetricsHelper) HTTPRegistrationDuration() prometheus.Histogram {
	return h.metrics.HTTPRegistrationDuration
}

func (h *MetricsHelper) HTTPUserLookupDuration() prometheus.Histogram {
	return h.metrics.HTTPUserLookupDuration
}

func (h *MetricsHelper) ErrorsTotal() *prometheus.CounterVec {
	return h.metrics.ErrorsTotal
}

func (h *MetricsHelper) SlowResponseTime() *prometheus.GaugeVec {
	return h.metrics.SlowResponseTime
}

func (h *MetricsHelper) ResponseTimePercentiles() *prometheus.HistogramVec {
	return h.metrics.ResponseTimePercentiles
}

func (h *MetricsHelper) GRPCConcurrentRequests() prometheus.Gauge {
	return h.metrics.GRPCConcurrentRequests
}

func (h *MetricsHelper) ValidationErrorsTotal() *prometheus.CounterVec {
	return h.metrics.ValidationErrorsTotal
}

func (h *MetricsHelper) AuthErrorsTotal() *prometheus.CounterVec {
	return h.metrics.AuthErrorsTotal
}

func (h *MetricsHelper) UserOperationsTotal() *prometheus.CounterVec {
	return h.metrics.UserOperationsTotal
}

func (h *MetricsHelper) UsersTotal() prometheus.Gauge {
	return h.metrics.UsersTotal
}

func (h *MetricsHelper) UsersByRole() *prometheus.GaugeVec {
	return h.metrics.UsersByRole
}

func (h *MetricsHelper) IPBlocksActive() *prometheus.GaugeVec {
	return h.metrics.IPBlocksActive
}

func (h *MetricsHelper) ActiveBruteforceAttacks() prometheus.Gauge {
	return h.metrics.ActiveBruteforceAttacks
}

func (h *MetricsHelper) AppHealthCheckDuration() prometheus.Histogram {
	return h.metrics.AppHealthCheckDuration
}

func (h *MetricsHelper) UpdateUptime() {
	h.metrics.UpdateUptime()
}

func (h *MetricsHelper) GetAppStartTime() time.Time {
	return h.metrics.GetAppStartTime()
}

func (h *MetricsHelper) GRPCStreamMessagesSent() prometheus.Counter {
	return h.metrics.GRPCStreamMessagesSent
}

func (h *MetricsHelper) GRPCStreamMessagesReceived() prometheus.Counter {
	return h.metrics.GRPCStreamMessagesReceived
}

func (h *MetricsHelper) DBUserQueriesTotal() *prometheus.CounterVec {
	return h.metrics.DBUserQueriesTotal
}

func (h *MetricsHelper) RedisIPBlockHitsTotal() prometheus.Counter {
	return h.metrics.RedisIPBlockHitsTotal
}

func (h *MetricsHelper) RedisIPBlockMissesTotal() prometheus.Counter {
	return h.metrics.RedisIPBlockMissesTotal
}

func (h *MetricsHelper) RedisTemporaryBlocksCreated() prometheus.Counter {
	return h.metrics.RedisTemporaryBlocksCreated
}

func (h *MetricsHelper) UsersActiveDaily() prometheus.Gauge {
	return h.metrics.UsersActiveDaily
}

func (h *MetricsHelper) UsersActiveWeekly() prometheus.Gauge {
	return h.metrics.UsersActiveWeekly
}

func (h *MetricsHelper) PeakUsageHours() *prometheus.GaugeVec {
	return h.metrics.PeakUsageHours
}

func (h *MetricsHelper) UserGeographicDistribution() *prometheus.GaugeVec {
	return h.metrics.UserGeographicDistribution
}

func (h *MetricsHelper) HighErrorRate() *prometheus.GaugeVec {
	return h.metrics.HighErrorRate
}

func (h *MetricsHelper) DiskSpaceLow() *prometheus.GaugeVec {
	return h.metrics.DiskSpaceLow
}

func (h *MetricsHelper) UnusualTrafficPatterns() prometheus.Gauge {
	return h.metrics.UnusualTrafficPatterns
}

func (h *MetricsHelper) PotentialDDOSDetected() prometheus.Gauge {
	return h.metrics.PotentialDDOSDetected
}
