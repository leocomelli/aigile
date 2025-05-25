package main

import (
	"log"
	"log/slog"

	"github.com/leocomelli/aigile/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		slog.Error("failed to execute command", "error", err)
		log.Fatal(err)
	}
}
