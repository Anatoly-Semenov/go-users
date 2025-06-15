package cli

import (
	"database/sql"
	"log"

	"github.com/anatoly_dev/go-users/internal/config"
	"github.com/anatoly_dev/go-users/internal/infrastructure/repository"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  `Run database migrations to setup the required schema`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.NewConfig(cmd)
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		db, err := sql.Open("postgres", cfg.GetDSN())
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()

		if err := db.Ping(); err != nil {
			log.Fatalf("Failed to ping database: %v", err)
		}

		migrationsPath := "migrations"
		if pathFlag, err := cmd.Flags().GetString("path"); err == nil && pathFlag != "" {
			migrationsPath = pathFlag
		}
		
		log.Printf("Applying migrations from %s", migrationsPath)
		if err := repository.MigrateDB(db, migrationsPath); err != nil {
			log.Fatalf("Failed to apply migrations: %v", err)
		}

		log.Println("Migrations applied successfully")
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().StringP("path", "p", "migrations", "Path to migrations directory")
	migrateCmd.Flags().StringP("config", "c", "", "Path to config file")
}
