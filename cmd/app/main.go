package main

import (
	"log"
	"os"

	"github.com/anatoly_dev/go-users/cmd/app/cli"
	"github.com/anatoly_dev/go-users/internal/config"
	"github.com/anatoly_dev/go-users/pkg/logger"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type Application struct {
	cmd *cobra.Command
}

func NewApplication() *Application {
	return &Application{
		cmd: &cobra.Command{},
	}
}

func (a *Application) Initialize() error {
	cfg, err := config.NewConfig(a.cmd)
	if err != nil {
		return err
	}

	if err := logger.Initialize(cfg); err != nil {
		return err
	}

	return nil
}

func (a *Application) Run() error {
	return cli.Execute()
}

func main() {
	app := NewApplication()

	if err := app.Initialize(); err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
		os.Exit(1)
	}

	if err := app.Run(); err != nil {
		logger.Fatal("Failed to execute command", zap.Error(err))
	}
}
