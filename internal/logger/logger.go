package logger

import (
	"io"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

func New(logFile string) *slog.Logger {
	writer := io.MultiWriter(os.Stdout, &lumberjack.Logger{Filename: logFile, MaxSize: 10, MaxBackups: 5, Compress: false})
	handler := slog.NewTextHandler(writer, &slog.HandlerOptions{Level: slog.LevelInfo})
	return slog.New(handler)
}
