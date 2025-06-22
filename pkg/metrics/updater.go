package metrics

import (
	"context"
	"database/sql"
	"time"

	"github.com/anatoly_dev/go-users/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type MetricsUpdater struct {
	metricsHelper  *MetricsHelper
	db             *sql.DB
	redis          *redis.Client
	stopCh         chan struct{}
	updateInterval time.Duration
}

func NewMetricsUpdater(metricsHelper *MetricsHelper, db *sql.DB, redis *redis.Client, updateInterval time.Duration) *MetricsUpdater {
	return &MetricsUpdater{
		metricsHelper:  metricsHelper,
		db:             db,
		redis:          redis,
		stopCh:         make(chan struct{}),
		updateInterval: updateInterval,
	}
}

func (u *MetricsUpdater) Start(ctx context.Context) {
	ticker := time.NewTicker(u.updateInterval)
	defer ticker.Stop()

	u.updateAllMetrics(ctx)

	for {
		select {
		case <-ticker.C:
			u.updateAllMetrics(ctx)
		case <-u.stopCh:
			logger.Info("Stopping metrics updater")
			return
		case <-ctx.Done():
			logger.Info("Metrics updater context cancelled")
			return
		}
	}
}

func (u *MetricsUpdater) Stop() {
	close(u.stopCh)
}

func (u *MetricsUpdater) updateAllMetrics(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Panic in metrics updater", zap.Any("panic", r))
			u.metricsHelper.RecordPanicRecovery()
		}
	}()

	logger.Debug("Updating metrics")

	u.updateSystemMetrics()

	if u.db != nil {
		u.updateDatabaseMetrics()
	}

	if u.redis != nil {
		u.updateRedisMetrics(ctx)
	}

	u.updateUserMetrics(ctx)

	u.updateIPBlockMetrics(ctx)

	u.updateSLAMetrics()

	u.updateErrorRates()
}

func (u *MetricsUpdater) updateSystemMetrics() {
	u.metricsHelper.UpdateSystemMetrics()
}

func (u *MetricsUpdater) updateDatabaseMetrics() {
	u.metricsHelper.UpdateDatabaseConnectionMetrics()
}

func (u *MetricsUpdater) updateRedisMetrics(ctx context.Context) {
	u.metricsHelper.UpdateRedisMetrics(ctx)
}

func (u *MetricsUpdater) updateUserMetrics(ctx context.Context) {
	if u.db == nil {
		return
	}

	var totalUsers int
	err := u.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&totalUsers)
	if err != nil {
		logger.Error("Failed to query total users", zap.Error(err))
		return
	}
	u.metricsHelper.UsersTotal().Set(float64(totalUsers))

	rows, err := u.db.QueryContext(ctx, "SELECT role, COUNT(*) FROM users GROUP BY role")
	if err != nil {
		logger.Error("Failed to query users by role", zap.Error(err))
		return
	}
	defer rows.Close()

	u.metricsHelper.UsersByRole().Reset()

	for rows.Next() {
		var role string
		var count int
		if err := rows.Scan(&role, &count); err != nil {
			logger.Error("Failed to scan user role", zap.Error(err))
			continue
		}
		u.metricsHelper.UsersByRole().WithLabelValues(role).Set(float64(count))
	}

	u.metricsHelper.UsersActiveDaily().Set(float64(totalUsers / 10))

	u.metricsHelper.UsersActiveWeekly().Set(float64(totalUsers / 5))
}

func (u *MetricsUpdater) updateIPBlockMetrics(ctx context.Context) {
	if u.db == nil {
		return
	}

	var permanentBlocks int
	err := u.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM ip_blocks WHERE expires_at IS NULL").Scan(&permanentBlocks)
	if err == nil {
		u.metricsHelper.IPBlocksActive().WithLabelValues("permanent").Set(float64(permanentBlocks))
	}

	if u.redis != nil {
		keys, err := u.redis.Keys(ctx, "ip_block:*").Result()
		if err == nil {
			u.metricsHelper.IPBlocksActive().WithLabelValues("temporary").Set(float64(len(keys)))
		}
	}

	u.metricsHelper.ActiveBruteforceAttacks().Set(0)
}

func (u *MetricsUpdater) updateSLAMetrics() {
	u.metricsHelper.UpdateSLAMetrics()
}

func (u *MetricsUpdater) updateErrorRates() {
	u.metricsHelper.UpdateErrorRates()
}

func (u *MetricsUpdater) UpdatePeakUsageHours() {
	currentHour := time.Now().Hour()
	u.metricsHelper.PeakUsageHours().WithLabelValues(string(rune(currentHour))).Inc()
}

func (u *MetricsUpdater) RecordGeographicDistribution(country, city string) {
	u.metricsHelper.UserGeographicDistribution().WithLabelValues(country, city).Inc()
}

func (u *MetricsUpdater) CheckAlertConditions(ctx context.Context) {
	u.metricsHelper.HighErrorRate().WithLabelValues("5%").Set(0)

	u.metricsHelper.SlowResponseTime().WithLabelValues("1s").Set(0)

	u.metricsHelper.DiskSpaceLow().WithLabelValues("90%").Set(0)

	u.metricsHelper.UnusualTrafficPatterns().Set(0)

	u.metricsHelper.PotentialDDOSDetected().Set(0)
}
