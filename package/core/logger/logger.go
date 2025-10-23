package logger

import (
	"log/slog"
	"os"

	"github.com/golang-cz/devslog"
)

func SetupLogger() {
	logger := slog.New(devslog.NewHandler(os.Stdout, nil))

	// optional: set global logger
	slog.SetDefault(logger)
}
