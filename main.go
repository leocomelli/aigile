// Package main is the entry point for the aigile CLI application.
package main

import (
	"log"
	"log/slog"

	"github.com/leocomelli/aigile/cmd"
)

// main is the entry point for the aigile CLI application.
func main() {
	if err := cmd.Execute(); err != nil {
		slog.Error("failed to execute command", "error", err)
		log.Fatal(err)
	}
}
