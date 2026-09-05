package logger

import (
	"gopkg.in/lumberjack.v2"
	"io"
	"log/slog"
	"os"
)

func New() *slog.Logger {

	rotator := &lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   false,
	}

	writer := io.MultiWriter(os.Stdout, rotator)

	return slog.New(slog.NewTextHandler(writer, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}
