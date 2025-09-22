package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	LogLevel    string
	Environment string
}

type ServerConfig struct {
	HTTPPort string
	GRPCPort string
	Host     string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	SecretKey     string
	TokenDuration time.Duration
}

func NewConfig(cmd *cobra.Command) (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Host: getEnv("HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "users"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			SecretKey:     getEnv("JWT_SECRET_KEY", "your-secret-key"),
			TokenDuration: time.Duration(getEnvAsInt("JWT_EXPIRATION_HOURS", 24)) * time.Hour,
		},
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		Environment: getEnv("ENVIRONMENT", "development"),
	}

	httpPort, err := cmd.Flags().GetString("port")
	if err != nil {
		return nil, err
	}
	if httpPort == "" {
		httpPort = getEnv("HTTP_PORT", "8080")
	}
	config.Server.HTTPPort = httpPort

	grpcPort, err := cmd.Flags().GetString("port")
	if err != nil {
		return nil, err
	}
	if grpcPort == "" {
		grpcPort = getEnv("GRPC_PORT", "50051")
	}
	config.Server.GRPCPort = grpcPort

	return config, nil
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}
