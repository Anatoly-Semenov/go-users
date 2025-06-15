package cli

import (
	"database/sql"
	"log"

	"github.com/anatoly_dev/go-users/internal/app"
	"github.com/anatoly_dev/go-users/internal/config"
	authimpl "github.com/anatoly_dev/go-users/internal/infrastructure/auth"
	"github.com/anatoly_dev/go-users/internal/infrastructure/repository"
	grpcserver "github.com/anatoly_dev/go-users/internal/infrastructure/server/grpc"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
)

var grpcServerCmd = &cobra.Command{
	Use:   "grpc-server",
	Short: "Start gRPC server",
	Long:  `Start the gRPC server for the user management service`,
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

		server := grpcserver.NewServer(userService, cfg)

		log.Printf("Starting gRPC server on port %s", cfg.Server.GRPCPort)
		if err := server.Start(); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	},
}

func init() {
	grpcServerCmd.Flags().StringP("port", "p", "50051", "Port to run the gRPC server on")
	grpcServerCmd.Flags().StringP("config", "c", "", "Path to config file")
}
