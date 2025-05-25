package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var (
	logLevel string
	rootCmd  = &cobra.Command{
		Use:   "aigile",
		Short: "A tool to generate Epics, Features, User Stories and Tasks",
		Long:  `Aigile is a CLI tool that helps you generate Epics, Features, User Stories and Tasks using LLMs (OpenAI, Gemini, Azure OpenAI) and integrates with GitHub Projects or Azure DevOps.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: GetLogLevel(),
			}))
			slog.SetDefault(logger)
			slog.Info("starting aigile", "log_level", logLevel)
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "Set log level (debug, info, warn, error)")
}

// GetLogLevel returns the slog.Level based on the command line flag
func GetLogLevel() slog.Level {
	switch logLevel {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func Execute() error {
	return rootCmd.Execute()
}
