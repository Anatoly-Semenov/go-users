package cli

import (
	"github.com/spf13/cobra"
)

// CommandManager управляет командами CLI приложения
type CommandManager struct {
	rootCmd *cobra.Command
}

// NewCommandManager создает новый менеджер команд
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

// initCommands инициализирует и добавляет все команды
func (cm *CommandManager) initCommands() {
	// Команды будут добавлены из соответствующих файлов через init()
}

// AddCommand добавляет команду к корневой команде
func (cm *CommandManager) AddCommand(cmds ...*cobra.Command) {
	cm.rootCmd.AddCommand(cmds...)
}

// Execute выполняет корневую команду
func (cm *CommandManager) Execute() error {
	return cm.rootCmd.Execute()
}

// Singleton-инстанс для доступа из других файлов пакета
var defaultManager *CommandManager

// init инициализирует синглтон
func init() {
	defaultManager = NewCommandManager()
}

// Execute предоставляет доступ к выполнению команд через синглтон
func Execute() error {
	return defaultManager.Execute()
}

// GetRootCmd возвращает корневую команду для использования в других файлах
func GetRootCmd() *cobra.Command {
	return defaultManager.rootCmd
}
