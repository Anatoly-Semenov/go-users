package cli

import (
	"database/sql"
	"log"

	"github.com/anatoly_dev/go-users/internal/config"
	"github.com/anatoly_dev/go-users/internal/infrastructure/repository"
	"github.com/anatoly_dev/go-users/pkg/logger"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type MigrateCommand struct {
	cmd *cobra.Command
}

func newMigrateCommand() *cobra.Command {
	migrateCmd := &MigrateCommand{}
	migrateCmd.cmd = &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		Long:  `Run database migrations to setup the required schema`,
		RunE:  migrateCmd.run,
	}

	migrateCmd.setupFlags()

	return migrateCmd.cmd
}

func (m *MigrateCommand) setupFlags() {
	m.cmd.Flags().StringP("path", "p", "migrations", "Path to migrations directory")
	m.cmd.Flags().StringP("config", "c", "", "Path to config file")
}

func (m *MigrateCommand) run(cmd *cobra.Command, args []string) error {
	cfg, err := m.initConfig(cmd)
	if err != nil {
		return err
	}

	db, err := m.connectDatabase(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	return m.runMigrations(cmd, db)
}

func (m *MigrateCommand) initConfig(cmd *cobra.Command) (*config.Config, error) {
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

func (m *MigrateCommand) connectDatabase(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
		return nil, err
	}

	if err := db.Ping(); err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
		return nil, err
	}

	return db, nil
}

func (m *MigrateCommand) runMigrations(cmd *cobra.Command, db *sql.DB) error {
	migrationsPath := "migrations"
	if pathFlag, err := cmd.Flags().GetString("path"); err == nil && pathFlag != "" {
		migrationsPath = pathFlag
	}

	logger.Info("Applying migrations", zap.String("path", migrationsPath))
	if err := repository.MigrateDB(db, migrationsPath); err != nil {
		logger.Fatal("Failed to apply migrations", zap.Error(err))
		return err
	}

	logger.Info("Migrations applied successfully")
	return nil
}

func init() {
	defaultManager.AddCommand(newMigrateCommand())
}
