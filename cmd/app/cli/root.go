package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "users",
	Short: "User management service",
	Long:  `A user management service with authentication and authorization capabilities`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(httpServerCmd)
	rootCmd.AddCommand(grpcServerCmd)
	rootCmd.AddCommand(migrateCmd)
}
