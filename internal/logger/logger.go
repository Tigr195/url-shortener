package logger

import (
	"io"
	"log/slog"
	"os"
)

func New() *slog.Logger {
	// создаём папку для логов если нет
	if err := os.MkdirAll("logs", 0755); err != nil {
		panic("failed to create logs dir: " + err.Error())
	}

	// открываем файл логов (append режим)
	file, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic("failed to open log file: " + err.Error())
	}

	// пишем и в файл и в консоль одновременно
	writer := io.MultiWriter(os.Stdout, file)

	return slog.New(slog.NewTextHandler(writer, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}
