package cli

import (
	"github.com/spf13/cobra"
)

type CommandManager struct {
	rootCmd *cobra.Command
}

func NewCommandManager() *CommandManager {
	cm := &CommandManager{
		rootCmd: &cobra.Command{
			Use:   "users",
			Short: "User management service",
			Long:  `A user management service with authentication and authorization capabilities`,
		},
	}

	cm.initCommands()

	return cm
}

func (cm *CommandManager) initCommands() {
}

func (cm *CommandManager) AddCommand(cmds ...*cobra.Command) {
	cm.rootCmd.AddCommand(cmds...)
}

func (cm *CommandManager) Execute() error {
	return cm.rootCmd.Execute()
}

var defaultManager *CommandManager

func init() {
	defaultManager = NewCommandManager()
}

func Execute() error {
	return defaultManager.Execute()
}

func GetRootCmd() *cobra.Command {
	return defaultManager.rootCmd
}
