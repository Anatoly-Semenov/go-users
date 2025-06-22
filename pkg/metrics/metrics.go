package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	UsersTotal             prometheus.Gauge
	UsersRegisteredTotal   prometheus.Counter
	UsersActiveDaily       prometheus.Gauge
	UsersActiveWeekly      prometheus.Gauge
	UsersByRole            *prometheus.GaugeVec
	UserOperationsTotal    *prometheus.CounterVec
	UserLoginAttemptsTotal *prometheus.CounterVec
	UserLoginSuccessRate   prometheus.Gauge

	JWTTokensIssuedTotal       prometheus.Counter
	JWTTokensValidatedTotal    *prometheus.CounterVec
	JWTTokenValidationDuration prometheus.Histogram
	PasswordResetRequestsTotal prometheus.Counter

	IPBlocksTotal        *prometheus.CounterVec
	IPBlocksActive       *prometheus.GaugeVec
	IPBlocksCreatedTotal *prometheus.CounterVec
	IPBlocksExpiredTotal prometheus.Counter
	IPBlocksRemovedTotal prometheus.Counter

	BruteforceAttemptsTotal  *prometheus.CounterVec
	BruteforceDetectedTotal  prometheus.Counter
	FailedLoginAttemptsTotal *prometheus.CounterVec
	BlockedRequestsTotal     *prometheus.CounterVec
	SuspiciousActivityTotal  *prometheus.CounterVec
	SecurityViolationsTotal  *prometheus.CounterVec

	HTTPRequestsTotal        *prometheus.CounterVec
	HTTPRequestDuration      *prometheus.HistogramVec
	HTTPRequestSizeBytes     *prometheus.HistogramVec
	HTTPResponseSizeBytes    *prometheus.HistogramVec
	HTTPConcurrentRequests   prometheus.Gauge
	HTTPLoginDuration        prometheus.Histogram
	HTTPRegistrationDuration prometheus.Histogram
	HTTPUserLookupDuration   prometheus.Histogram

	GRPCRequestsTotal          *prometheus.CounterVec
	GRPCRequestDuration        *prometheus.HistogramVec
	GRPCRequestSizeBytes       *prometheus.HistogramVec
	GRPCResponseSizeBytes      *prometheus.HistogramVec
	GRPCConcurrentRequests     prometheus.Gauge
	GRPCStreamMessagesSent     prometheus.Counter
	GRPCStreamMessagesReceived prometheus.Counter
	GRPCClientConnections      prometheus.Gauge

	DBConnectionsActive      prometheus.Gauge
	DBConnectionsIdle        prometheus.Gauge
	DBConnectionsMax         prometheus.Gauge
	DBConnectionWaitDuration prometheus.Histogram
	DBOperationsTotal        *prometheus.CounterVec
	DBOperationDuration      *prometheus.HistogramVec
	DBRowsAffectedTotal      *prometheus.CounterVec
	DBQueryErrorsTotal       *prometheus.CounterVec
	DBUserQueriesTotal       *prometheus.CounterVec
	DBMigrationDuration      prometheus.Histogram

	RedisConnectionsActive      prometheus.Gauge
	RedisCommandsTotal          *prometheus.CounterVec
	RedisCommandDuration        *prometheus.HistogramVec
	RedisMemoryUsedBytes        prometheus.Gauge
	RedisKeysTotal              *prometheus.GaugeVec
	RedisKeysExpiredTotal       prometheus.Counter
	RedisIPBlocksTotal          prometheus.Gauge
	RedisIPBlockHitsTotal       prometheus.Counter
	RedisIPBlockMissesTotal     prometheus.Counter
	RedisTemporaryBlocksCreated prometheus.Counter
	RedisTemporaryBlocksExpired prometheus.Counter

	AppUptimeSeconds          prometheus.Gauge
	AppRestartTotal           prometheus.Counter
	AppConfigReloadsTotal     prometheus.Counter
	AppHealthCheckDuration    prometheus.Histogram
	MiddlewareRequestDuration *prometheus.HistogramVec
	MiddlewareErrorsTotal     *prometheus.CounterVec
	MiddlewareBypassed        *prometheus.CounterVec

	ErrorsTotal            *prometheus.CounterVec
	ErrorRate              *prometheus.GaugeVec
	PanicRecoveriesTotal   prometheus.Counter
	AuthErrorsTotal        *prometheus.CounterVec
	ValidationErrorsTotal  *prometheus.CounterVec
	DBConstraintViolations *prometheus.CounterVec

	ServiceAvailabilityPercentage prometheus.Gauge
	EndpointAvailability          *prometheus.GaugeVec
	ResponseTimePercentiles       *prometheus.HistogramVec
	ErrorBudgetRemaining          prometheus.Gauge

	HighErrorRate           *prometheus.GaugeVec
	SlowResponseTime        *prometheus.GaugeVec
	DBConnectionExhaustion  prometheus.Gauge
	MemoryUsageHigh         *prometheus.GaugeVec
	DiskSpaceLow            *prometheus.GaugeVec
	ActiveBruteforceAttacks prometheus.Gauge
	UnusualTrafficPatterns  prometheus.Gauge
	PotentialDDOSDetected   prometheus.Gauge

	UserSessionDuration        prometheus.Histogram
	UserActionsPerSession      prometheus.Histogram
	PopularEndpoints           *prometheus.CounterVec
	UserGeographicDistribution *prometheus.GaugeVec
	PeakUsageHours             *prometheus.GaugeVec

	appStartTime time.Time
}

func NewMetrics() *Metrics {
	return &Metrics{
		appStartTime: time.Now(),

		UsersTotal: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "users_total_count",
			Help: "Total number of users in the system",
		}),

		UsersRegisteredTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "users_registered_total",
			Help: "Total number of registered users",
		}),

		UsersActiveDaily: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "users_active_daily",
			Help: "Number of active users in the last 24 hours",
		}),

		UsersActiveWeekly: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "users_active_weekly",
			Help: "Number of active users in the last week",
		}),

		UsersByRole: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "users_by_role",
			Help: "Number of users by role",
		}, []string{"role"}),

		UserOperationsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "user_operations_total",
			Help: "Total number of user operations",
		}, []string{"operation", "status"}),

		UserLoginAttemptsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "user_login_attempts_total",
			Help: "Total number of login attempts",
		}, []string{"status"}),

		UserLoginSuccessRate: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "user_login_success_rate",
			Help: "Login success rate percentage",
		}),

		JWTTokensIssuedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "jwt_tokens_issued_total",
			Help: "Total number of JWT tokens issued",
		}),

		JWTTokensValidatedTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "jwt_tokens_validated_total",
			Help: "Total number of JWT token validations",
		}, []string{"status"}),

		JWTTokenValidationDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "jwt_token_validation_duration_seconds",
			Help:    "Time spent validating JWT tokens",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		}),

		PasswordResetRequestsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "password_reset_requests_total",
			Help: "Total number of password reset requests",
		}),

		IPBlocksTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "ip_blocks_total",
			Help: "Total number of IP blocks",
		}, []string{"type"}),

		IPBlocksActive: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ip_blocks_active",
			Help: "Number of active IP blocks",
		}, []string{"type"}),

		IPBlocksCreatedTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "ip_blocks_created_total",
			Help: "Total number of IP blocks created",
		}, []string{"type", "reason"}),

		IPBlocksExpiredTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "ip_blocks_expired_total",
			Help: "Total number of expired IP blocks",
		}),

		IPBlocksRemovedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "ip_blocks_removed_total",
			Help: "Total number of removed IP blocks",
		}),

		BruteforceAttemptsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "bruteforce_attempts_total",
			Help: "Total number of brute force attempts",
		}, []string{"ip"}),

		BruteforceDetectedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "bruteforce_detected_total",
			Help: "Total number of detected brute force attacks",
		}),

		FailedLoginAttemptsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "failed_login_attempts_total",
			Help: "Total number of failed login attempts",
		}, []string{"ip"}),

		BlockedRequestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "blocked_requests_total",
			Help: "Total number of blocked requests",
		}, []string{"reason"}),

		SuspiciousActivityTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "suspicious_activity_detected_total",
			Help: "Total number of suspicious activities detected",
		}, []string{"type"}),

		SecurityViolationsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "security_violations_total",
			Help: "Total number of security violations",
		}, []string{"type"}),

		HTTPRequestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		}, []string{"method", "endpoint", "status"}),

		HTTPRequestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
		}, []string{"method", "endpoint", "status"}),

		HTTPRequestSizeBytes: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 7),
		}, []string{"method", "endpoint"}),

		HTTPResponseSizeBytes: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 7),
		}, []string{"method", "endpoint"}),

		HTTPConcurrentRequests: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "http_concurrent_requests",
			Help: "Number of concurrent HTTP requests",
		}),

		HTTPLoginDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "http_login_duration_seconds",
			Help:    "HTTP login request duration",
			Buckets: []float64{0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
		}),

		HTTPRegistrationDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "http_registration_duration_seconds",
			Help:    "HTTP registration request duration",
			Buckets: []float64{0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
		}),

		HTTPUserLookupDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "http_user_lookup_duration_seconds",
			Help:    "HTTP user lookup duration",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5},
		}),

		GRPCRequestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		}, []string{"method", "status"}),

		GRPCRequestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
		}, []string{"method", "status"}),

		GRPCRequestSizeBytes: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "grpc_request_size_bytes",
			Help:    "gRPC request size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 7),
		}, []string{"method"}),

		GRPCResponseSizeBytes: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "grpc_response_size_bytes",
			Help:    "gRPC response size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 7),
		}, []string{"method"}),

		GRPCConcurrentRequests: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "grpc_concurrent_requests",
			Help: "Number of concurrent gRPC requests",
		}),

		GRPCStreamMessagesSent: promauto.NewCounter(prometheus.CounterOpts{
			Name: "grpc_stream_messages_sent_total",
			Help: "Total number of gRPC stream messages sent",
		}),

		GRPCStreamMessagesReceived: promauto.NewCounter(prometheus.CounterOpts{
			Name: "grpc_stream_messages_received_total",
			Help: "Total number of gRPC stream messages received",
		}),

		GRPCClientConnections: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "grpc_client_connections_total",
			Help: "Number of gRPC client connections",
		}),

		DBConnectionsActive: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections",
		}),

		DBConnectionsIdle: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "db_connections_idle",
			Help: "Number of idle database connections",
		}),

		DBConnectionsMax: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "db_connections_max",
			Help: "Maximum number of database connections",
		}),

		DBConnectionWaitDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "db_connection_wait_duration_seconds",
			Help:    "Time spent waiting for database connection",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
		}),

		DBOperationsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "db_operations_total",
			Help: "Total number of database operations",
		}, []string{"operation", "table"}),

		DBOperationDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "db_operation_duration_seconds",
			Help:    "Database operation duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
		}, []string{"operation", "table"}),

		DBRowsAffectedTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "db_rows_affected_total",
			Help: "Total number of database rows affected",
		}, []string{"operation", "table"}),

		DBQueryErrorsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "db_query_errors_total",
			Help: "Total number of database query errors",
		}, []string{"error_type"}),

		DBUserQueriesTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "db_user_queries_total",
			Help: "Total number of user-related database queries",
		}, []string{"type"}),

		DBMigrationDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "db_migration_duration_seconds",
			Help:    "Database migration duration in seconds",
			Buckets: []float64{1, 5, 10, 30, 60, 120, 300, 600},
		}),

		RedisConnectionsActive: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "redis_connections_active",
			Help: "Number of active Redis connections",
		}),

		RedisCommandsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "redis_commands_total",
			Help: "Total number of Redis commands executed",
		}, []string{"command"}),

		RedisCommandDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "redis_command_duration_seconds",
			Help:    "Redis command duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		}, []string{"command"}),

		RedisMemoryUsedBytes: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "redis_memory_used_bytes",
			Help: "Redis memory usage in bytes",
		}),

		RedisKeysTotal: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "redis_keys_total",
			Help: "Total number of Redis keys",
		}, []string{"type"}),

		RedisKeysExpiredTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "redis_keys_expired_total",
			Help: "Total number of expired Redis keys",
		}),

		RedisIPBlocksTotal: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "redis_ip_blocks_total",
			Help: "Total number of IP blocks in Redis",
		}),

		RedisIPBlockHitsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "redis_ip_block_hits_total",
			Help: "Total number of IP block cache hits",
		}),

		RedisIPBlockMissesTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "redis_ip_block_misses_total",
			Help: "Total number of IP block cache misses",
		}),

		RedisTemporaryBlocksCreated: promauto.NewCounter(prometheus.CounterOpts{
			Name: "redis_temporary_blocks_created_total",
			Help: "Total number of temporary blocks created in Redis",
		}),

		RedisTemporaryBlocksExpired: promauto.NewCounter(prometheus.CounterOpts{
			Name: "redis_temporary_blocks_expired_total",
			Help: "Total number of temporary blocks expired in Redis",
		}),

		AppUptimeSeconds: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "app_uptime_seconds",
			Help: "Application uptime in seconds",
		}),

		AppRestartTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "app_restart_total",
			Help: "Total number of application restarts",
		}),

		AppConfigReloadsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "app_config_reloads_total",
			Help: "Total number of configuration reloads",
		}),

		AppHealthCheckDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "app_health_check_duration_seconds",
			Help:    "Application health check duration",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		}),

		MiddlewareRequestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "middleware_request_duration_seconds",
			Help:    "Middleware processing duration",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5},
		}, []string{"middleware"}),

		MiddlewareErrorsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "middleware_processing_errors_total",
			Help: "Total number of middleware processing errors",
		}, []string{"middleware", "error_type"}),

		MiddlewareBypassed: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "middleware_bypassed_total",
			Help: "Total number of middleware bypasses",
		}, []string{"middleware", "reason"}),

		ErrorsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Total number of errors",
		}, []string{"component", "error_type"}),

		ErrorRate: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "error_rate",
			Help: "Error rate by component",
		}, []string{"component"}),

		PanicRecoveriesTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "panic_recoveries_total",
			Help: "Total number of panic recoveries",
		}),

		AuthErrorsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "auth_errors_total",
			Help: "Total number of authentication errors",
		}, []string{"type"}),

		ValidationErrorsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "validation_errors_total",
			Help: "Total number of validation errors",
		}, []string{"field"}),

		DBConstraintViolations: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "db_constraint_violations_total",
			Help: "Total number of database constraint violations",
		}, []string{"constraint"}),

		ServiceAvailabilityPercentage: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "service_availability_percentage",
			Help: "Service availability percentage",
		}),

		EndpointAvailability: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "endpoint_availability",
			Help: "Endpoint availability",
		}, []string{"endpoint"}),

		ResponseTimePercentiles: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "response_time_percentiles",
			Help:    "Response time percentiles",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
		}, []string{"percentile"}),

		ErrorBudgetRemaining: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "error_budget_remaining",
			Help: "Remaining error budget",
		}),

		HighErrorRate: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "high_error_rate",
			Help: "High error rate indicator",
		}, []string{"threshold"}),

		SlowResponseTime: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "slow_response_time",
			Help: "Slow response time indicator",
		}, []string{"threshold"}),

		DBConnectionExhaustion: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "database_connection_exhaustion",
			Help: "Database connection exhaustion indicator",
		}),

		MemoryUsageHigh: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "memory_usage_high",
			Help: "High memory usage indicator",
		}, []string{"threshold"}),

		DiskSpaceLow: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "disk_space_low",
			Help: "Low disk space indicator",
		}, []string{"threshold"}),

		ActiveBruteforceAttacks: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "active_bruteforce_attacks",
			Help: "Number of active brute force attacks",
		}),

		UnusualTrafficPatterns: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "unusual_traffic_patterns",
			Help: "Unusual traffic patterns indicator",
		}),

		PotentialDDOSDetected: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "potential_ddos_detected",
			Help: "Potential DDoS attack indicator",
		}),

		UserSessionDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "user_session_duration_seconds",
			Help:    "User session duration in seconds",
			Buckets: []float64{60, 300, 900, 1800, 3600, 7200, 14400, 28800},
		}),

		UserActionsPerSession: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "user_actions_per_session",
			Help:    "Number of user actions per session",
			Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500},
		}),

		PopularEndpoints: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "popular_endpoints",
			Help: "Popular endpoints counter",
		}, []string{"endpoint"}),

		UserGeographicDistribution: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "user_geographic_distribution",
			Help: "User geographic distribution",
		}, []string{"country", "city"}),

		PeakUsageHours: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "peak_usage_hours",
			Help: "Peak usage hours",
		}, []string{"hour"}),
	}
}

func (m *Metrics) UpdateUptime() {
	m.AppUptimeSeconds.Set(time.Since(m.appStartTime).Seconds())
}

func (m *Metrics) GetAppStartTime() time.Time {
	return m.appStartTime
}
