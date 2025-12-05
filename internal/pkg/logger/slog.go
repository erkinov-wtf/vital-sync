package logger

import (
	"io"
	"log/slog"
	"os"

	"github.com/erkinov-wtf/vital-sync/internal/config"
	"github.com/erkinov-wtf/vital-sync/internal/constants"
)

func SetupLogger(env string) *slog.Logger {
	fw := &fileWriter{
		logDir: "logs",
	}

	var log *slog.Logger

	switch env {
	case config.LocalEnv:
		// use MultiWriter to write to both stdout and file
		mw := io.MultiWriter(os.Stdout, fw)
		handler := slog.NewTextHandler(mw, &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		})
		log = slog.New(handler)

	case config.ReleaseEnv:
		// use MultiWriter for JSON logging
		mw := io.MultiWriter(os.Stdout, fw)
		handler := slog.NewJSONHandler(mw, &slog.HandlerOptions{
			Level:     slog.LevelInfo,
			AddSource: true,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == "time" {
					return slog.String("time", a.Value.Time().Format(constants.LoggerFormat))
				}
				return a
			},
		})
		log = slog.New(handler)
	}

	return log
}
