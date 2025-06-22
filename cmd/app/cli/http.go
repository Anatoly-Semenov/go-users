package cli

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/anatoly_dev/go-users/internal/app"
	"github.com/anatoly_dev/go-users/internal/config"
	authimpl "github.com/anatoly_dev/go-users/internal/infrastructure/auth"
	"github.com/anatoly_dev/go-users/internal/infrastructure/repository"
	httpserver "github.com/anatoly_dev/go-users/internal/infrastructure/server/http"
	"github.com/anatoly_dev/go-users/pkg/logger"
	"github.com/anatoly_dev/go-users/pkg/metrics"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type HTTPServerCommand struct {
	cmd *cobra.Command
}

func newHTTPServerCommand() *cobra.Command {
	httpCmd := &HTTPServerCommand{}
	httpCmd.cmd = &cobra.Command{
		Use:   "http-server",
		Short: "Start HTTP server",
		Long:  `Start the HTTP server for the user management service`,
		RunE:  httpCmd.run,
	}

	httpCmd.setupFlags()

	return httpCmd.cmd
}

func (h *HTTPServerCommand) setupFlags() {
	h.cmd.Flags().StringP("port", "p", "8080", "Port to run the HTTP server on")
	h.cmd.Flags().StringP("config", "c", "", "Path to config file")
}

func (h *HTTPServerCommand) run(cmd *cobra.Command, args []string) error {
	cfg, err := h.initConfig(cmd)
	if err != nil {
		return err
	}

	db, err := h.connectDatabase(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	redisClient, err := h.connectRedis(cfg)
	if err != nil {
		return err
	}
	defer redisClient.Close()

	metricsSystem := metrics.NewMetrics()
	metricsHelper := metrics.NewMetricsHelper(metricsSystem, db, redisClient)

	metricsUpdater := metrics.NewMetricsUpdater(metricsHelper, db, redisClient, 30*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go metricsUpdater.Start(ctx)
	defer metricsUpdater.Stop()

	userService, handler, err := h.setupServices(db, redisClient, cfg, metricsHelper)
	if err != nil {
		return err
	}

	return h.startServer(userService, handler, cfg)
}

func (h *HTTPServerCommand) initConfig(cmd *cobra.Command) (*config.Config, error) {
	cfg, err := config.NewConfig(cmd)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
		return nil, err
	}

	if err := logger.Initialize(cfg); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
		return nil, err
	}

	return cfg, nil
}

func (h *HTTPServerCommand) connectDatabase(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
		return nil, err
	}

	if err := db.Ping(); err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
		return nil, err
	}

	start := time.Now()
	if err := repository.MigrateDB(db, "migrations"); err != nil {
		logger.Fatal("Failed to apply migrations", zap.Error(err))
		return nil, err
	}

	logger.Info("Database migrations completed",
		zap.Duration("duration", time.Since(start)))

	return db, nil
}

func (h *HTTPServerCommand) connectRedis(cfg *config.Config) (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
		return nil, err
	}

	logger.Info("Successfully connected to Redis", zap.String("addr", cfg.GetRedisAddr()))
	return redisClient, nil
}

func (h *HTTPServerCommand) setupServices(db *sql.DB, redisClient *redis.Client, cfg *config.Config, metricsHelper *metrics.MetricsHelper) (*app.UserService, http.Handler, error) {
	userRepo := repository.NewPostgresUserRepositoryWithMetrics(db, metricsHelper)

	ipBlockPostgresRepo := repository.NewPostgresIPBlockRepository(db)
	ipBlockRedisRepo := repository.NewRedisIPBlockRepositoryWithMetrics(redisClient, metricsHelper)

	bruteforceConfig := app.DefaultBruteforceDefenseConfig()
	ipBlockService := app.NewIPBlockService(ipBlockPostgresRepo, ipBlockRedisRepo, bruteforceConfig)

	baseAuthService := authimpl.NewJWTService(
		userRepo,
		cfg.JWT.SecretKey,
		cfg.JWT.TokenDuration,
	)

	securedAuthService := authimpl.NewSecuredAuthService(baseAuthService, ipBlockService)

	userService := app.NewUserService(userRepo, securedAuthService)

	userHandler := httpserver.NewUserHandler(userService, metricsHelper)

	metricsMiddleware := httpserver.NewMetricsMiddleware(metricsHelper)

	mux := http.NewServeMux()
	userHandler.RegisterRoutes(mux)

	var handler http.Handler = mux

	handler = metricsMiddleware.Middleware(handler)
	handler = metricsMiddleware.PanicRecoveryMiddleware(handler)
	handler = httpserver.SecurityMiddleware(metricsHelper)(handler)
	handler = httpserver.IPMiddleware(handler)
	handler = httpserver.IPBlockMiddleware(ipBlockService, metricsHelper)(handler)

	logger.Info("HTTP server services initialized with comprehensive metrics")
	return userService, handler, nil
}

func (h *HTTPServerCommand) startServer(userService *app.UserService, handler http.Handler, cfg *config.Config) error {
	server := &http.Server{
		Addr:    ":" + cfg.Server.HTTPPort,
		Handler: handler,
	}

	logger.Info("Starting HTTP server with comprehensive metrics and IP blocking protection",
		zap.String("port", cfg.Server.HTTPPort),
		zap.String("metrics_endpoint", "/metrics"),
		zap.String("health_endpoint", "/health"))

	return server.ListenAndServe()
}

func init() {
	defaultManager.AddCommand(newHTTPServerCommand())
}
