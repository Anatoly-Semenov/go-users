package cli

import (
	"database/sql"
	"log"

	"github.com/anatoly_dev/go-users/internal/app"
	"github.com/anatoly_dev/go-users/internal/config"
	authimpl "github.com/anatoly_dev/go-users/internal/infrastructure/auth"
	"github.com/anatoly_dev/go-users/internal/infrastructure/repository"
	httpserver "github.com/anatoly_dev/go-users/internal/infrastructure/server/http"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
)

var httpServerCmd = &cobra.Command{
	Use:   "http-server",
	Short: "Start HTTP server",
	Long:  `Start the HTTP server for the user management service`,
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

		if err := repository.MigrateDB(db, "migrations"); err != nil {
			log.Fatalf("Failed to apply migrations: %v", err)
		}

		userRepo := repository.NewPostgresUserRepository(db)

		authService := authimpl.NewJWTService(userRepo, cfg.JWT.SecretKey, cfg.JWT.TokenDuration)

		userService := app.NewUserService(userRepo, authService)

		server := httpserver.NewServer(userService, cfg)

		log.Printf("Starting HTTP server on port %s", cfg.Server.HTTPPort)
		if err := server.Start(); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	},
}

func init() {
	httpServerCmd.Flags().StringP("port", "p", "8080", "Port to run the HTTP server on")
	httpServerCmd.Flags().StringP("config", "c", "", "Path to config file")
}
