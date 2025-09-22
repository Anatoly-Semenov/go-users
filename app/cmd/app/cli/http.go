package cli

import (
	"context"
	"database/sql"
	app2 "github.com/anatoly_dev/go-users/app/internal/app"
	"github.com/anatoly_dev/go-users/app/internal/config"
	"github.com/anatoly_dev/go-users/app/internal/infrastructure/auth"
	repository2 "github.com/anatoly_dev/go-users/app/internal/infrastructure/repository"
	http2 "github.com/anatoly_dev/go-users/app/internal/infrastructure/server/http"
	"github.com/anatoly_dev/go-users/app/pkg/logger"
	metrics2 "github.com/anatoly_dev/go-users/app/pkg/metrics"
	"log"
	"net/http"
	"time"

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

	metricsSystem := metrics2.NewMetrics()
	metricsHelper := metrics2.NewMetricsHelper(metricsSystem, db, redisClient)

	metricsUpdater := metrics2.NewMetricsUpdater(metricsHelper, db, redisClient, 30*time.Second)
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
	if err := repository2.MigrateDB(db, "migrations"); err != nil {
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

func (h *HTTPServerCommand) setupServices(db *sql.DB, redisClient *redis.Client, cfg *config.Config, metricsHelper *metrics2.MetricsHelper) (*app2.UserService, http.Handler, error) {
	userRepo := repository2.NewPostgresUserRepositoryWithMetrics(db, metricsHelper)

	ipBlockPostgresRepo := repository2.NewPostgresIPBlockRepository(db)
	ipBlockRedisRepo := repository2.NewRedisIPBlockRepositoryWithMetrics(redisClient, metricsHelper)

	bruteforceConfig := app2.DefaultBruteforceDefenseConfig()
	ipBlockService := app2.NewIPBlockService(ipBlockPostgresRepo, ipBlockRedisRepo, bruteforceConfig)

	baseAuthService := auth.NewJWTService(
		userRepo,
		cfg.JWT.SecretKey,
		cfg.JWT.TokenDuration,
	)

	securedAuthService := auth.NewSecuredAuthService(baseAuthService, ipBlockService)

	userService := app2.NewUserService(userRepo, securedAuthService)

	userHandler := http2.NewUserHandler(userService, metricsHelper)

	metricsMiddleware := http2.NewMetricsMiddleware(metricsHelper)

	mux := http.NewServeMux()
	userHandler.RegisterRoutes(mux)

	var handler http.Handler = mux

	handler = metricsMiddleware.Middleware(handler)
	handler = metricsMiddleware.PanicRecoveryMiddleware(handler)
	handler = http2.SecurityMiddleware(metricsHelper)(handler)
	handler = http2.IPMiddleware(handler)
	handler = http2.IPBlockMiddleware(ipBlockService, metricsHelper)(handler)

	logger.Info("HTTP server services initialized with comprehensive metrics")
	return userService, handler, nil
}

func (h *HTTPServerCommand) startServer(userService *app2.UserService, handler http.Handler, cfg *config.Config) error {
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
