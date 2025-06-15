package cli

import (
	"database/sql"
	"log"

	"github.com/anatoly_dev/go-users/internal/app"
	"github.com/anatoly_dev/go-users/internal/config"
	authimpl "github.com/anatoly_dev/go-users/internal/infrastructure/auth"
	"github.com/anatoly_dev/go-users/internal/infrastructure/repository"
	httpserver "github.com/anatoly_dev/go-users/internal/infrastructure/server/http"
	"github.com/anatoly_dev/go-users/pkg/logger"
	_ "github.com/lib/pq"
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

	services, err := h.setupServices(db, cfg)
	if err != nil {
		return err
	}

	return h.startServer(services, cfg)
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

	if err := repository.MigrateDB(db, "migrations"); err != nil {
		logger.Fatal("Failed to apply migrations", zap.Error(err))
		return nil, err
	}

	return db, nil
}

func (h *HTTPServerCommand) setupServices(db *sql.DB, cfg *config.Config) (*app.UserService, error) {
	userRepo := repository.NewPostgresUserRepository(db)
	authService := authimpl.NewJWTService(userRepo, cfg.JWT.SecretKey, cfg.JWT.TokenDuration)
	userService := app.NewUserService(userRepo, authService)

	return userService, nil
}

func (h *HTTPServerCommand) startServer(userService *app.UserService, cfg *config.Config) error {
	server := httpserver.NewServer(userService, cfg)

	logger.Info("Starting HTTP server", zap.String("port", cfg.Server.HTTPPort))
	return server.Start()
}

func init() {
	defaultManager.AddCommand(newHTTPServerCommand())
}
