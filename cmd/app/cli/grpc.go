package cli

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/anatoly_dev/go-users/internal/app"
	"github.com/anatoly_dev/go-users/internal/config"
	authimpl "github.com/anatoly_dev/go-users/internal/infrastructure/auth"
	"github.com/anatoly_dev/go-users/internal/infrastructure/repository"
	grpcserver "github.com/anatoly_dev/go-users/internal/infrastructure/server/grpc"
	"github.com/anatoly_dev/go-users/pkg/logger"
	"github.com/anatoly_dev/go-users/pkg/metrics"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type GRPCServerCommand struct {
	cmd *cobra.Command
}

func newGRPCServerCommand() *cobra.Command {
	grpcCmd := &GRPCServerCommand{}
	grpcCmd.cmd = &cobra.Command{
		Use:   "grpc-server",
		Short: "Start gRPC server",
		Long:  `Start the gRPC server for the user management service`,
		RunE:  grpcCmd.run,
	}

	grpcCmd.setupFlags()

	return grpcCmd.cmd
}

func (g *GRPCServerCommand) setupFlags() {
	g.cmd.Flags().StringP("port", "p", "50051", "Port to run the gRPC server on")
	g.cmd.Flags().StringP("config", "c", "", "Path to config file")
}

func (g *GRPCServerCommand) run(cmd *cobra.Command, args []string) error {
	cfg, err := g.initConfig(cmd)
	if err != nil {
		return err
	}

	db, err := g.connectDatabase(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	redisClient, err := g.connectRedis(cfg)
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

	userService, server, err := g.setupServices(db, redisClient, cfg, metricsHelper)
	if err != nil {
		return err
	}

	return g.startServer(userService, server, cfg)
}

func (g *GRPCServerCommand) initConfig(cmd *cobra.Command) (*config.Config, error) {
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

func (g *GRPCServerCommand) connectDatabase(cfg *config.Config) (*sql.DB, error) {
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

func (g *GRPCServerCommand) connectRedis(cfg *config.Config) (*redis.Client, error) {
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

func (g *GRPCServerCommand) setupServices(db *sql.DB, redisClient *redis.Client, cfg *config.Config, metricsHelper *metrics.MetricsHelper) (*app.UserService, *grpcserver.Server, error) {
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

	metricsInterceptor := grpcserver.NewMetricsInterceptor(metricsHelper)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(metricsInterceptor.UnaryServerInterceptor()),
		grpc.StreamInterceptor(metricsInterceptor.StreamServerInterceptor()),
	)

	grpcServer := grpcserver.NewServerWithGRPCServer(userService, cfg, server)

	logger.Info("gRPC server services initialized with comprehensive metrics")
	return userService, grpcServer, nil
}

func (g *GRPCServerCommand) startServer(userService *app.UserService, server *grpcserver.Server, cfg *config.Config) error {
	logger.Info("Starting gRPC server with comprehensive metrics",
		zap.String("port", cfg.Server.GRPCPort))

	return server.Start()
}

func init() {
	defaultManager.AddCommand(newGRPCServerCommand())
}
