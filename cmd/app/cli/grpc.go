package cli

import (
	"database/sql"
	"log"

	"github.com/anatoly_dev/go-users/internal/app"
	"github.com/anatoly_dev/go-users/internal/config"
	authimpl "github.com/anatoly_dev/go-users/internal/infrastructure/auth"
	"github.com/anatoly_dev/go-users/internal/infrastructure/repository"
	grpcserver "github.com/anatoly_dev/go-users/internal/infrastructure/server/grpc"
	"github.com/anatoly_dev/go-users/pkg/logger"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
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

	services, err := g.setupServices(db, cfg)
	if err != nil {
		return err
	}

	return g.startServer(services, cfg)
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

	if err := repository.MigrateDB(db, "migrations"); err != nil {
		logger.Fatal("Failed to apply migrations", zap.Error(err))
		return nil, err
	}

	return db, nil
}

func (g *GRPCServerCommand) setupServices(db *sql.DB, cfg *config.Config) (*app.UserService, error) {
	userRepo := repository.NewPostgresUserRepository(db)
	authService := authimpl.NewJWTService(userRepo, cfg.JWT.SecretKey, cfg.JWT.TokenDuration)
	userService := app.NewUserService(userRepo, authService)

	return userService, nil
}

func (g *GRPCServerCommand) startServer(userService *app.UserService, cfg *config.Config) error {
	server := grpcserver.NewServer(userService, cfg)

	logger.Info("Starting gRPC server", zap.String("port", cfg.Server.GRPCPort))
	return server.Start()
}

func init() {
	defaultManager.AddCommand(newGRPCServerCommand())
}
