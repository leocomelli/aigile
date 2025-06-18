package cmd

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
)

// rootCmd is the base command for the aigile CLI application.
var (
	logLevel string
	rootCmd  = &cobra.Command{
		Use:   "aigile",
		Short: "A tool to generate User Stories and Tasks",
		Long:  `Aigile is a CLI tool that helps you generate User Stories and Tasks using LLMs (OpenAI, Gemini, Azure OpenAI) and integrates with GitHub Projects or Azure DevOps.`,
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			handler := tint.NewHandler(os.Stdout, &tint.Options{
				Level:      GetLogLevel(),
				TimeFormat: "15:04:05",
			})
			logger := slog.New(handler)
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

// Execute runs the root command for the CLI application.
func Execute() error {
	return rootCmd.Execute()
}
