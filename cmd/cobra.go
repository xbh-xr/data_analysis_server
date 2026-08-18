package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"go-admin/cmd/api"
	"go-admin/cmd/app"
	"go-admin/cmd/config"
	"go-admin/cmd/migrate"
	"go-admin/cmd/version"
)

var rootCmd = &cobra.Command{
	Use:               "go-admin",
	Short:             "go-admin",
	SilenceUsage:      true,
	Long:              `go-admin`,
	PersistentPreRunE: func(*cobra.Command, []string) error { return nil },
	RunE: func(cmd *cobra.Command, args []string) error {
		return api.Start()
	},
}

func init() {
	rootCmd.AddCommand(api.StartCmd)
	rootCmd.AddCommand(migrate.StartCmd)
	rootCmd.AddCommand(version.StartCmd)
	rootCmd.AddCommand(config.StartCmd)
	rootCmd.AddCommand(app.StartCmd)
}

// Execute : apply commands。无子命令时默认启动 API 主进程。
func Execute() {
	if !hasKnownCommand(os.Args[1:]) {
		rootCmd.SetArgs(append([]string{"server"}, os.Args[1:]...))
	}
	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

func hasKnownCommand(args []string) bool {
	if len(args) == 0 {
		return false
	}
	name := args[0]
	switch name {
	case "help", "completion", "-h", "--help":
		return true
	}
	if strings.HasPrefix(name, "-") {
		return false
	}
	for _, c := range rootCmd.Commands() {
		if c.Name() == name || c.HasAlias(name) {
			return true
		}
	}
	return false
}
