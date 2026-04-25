package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

func New(logFile string) *slog.Logger {
	var writer io.Writer = os.Stdout
	if strings.TrimSpace(logFile) != "" {
		writer = &lumberjack.Logger{Filename: logFile, MaxSize: 10, MaxBackups: 5, Compress: false}
	}
	handler := slog.NewTextHandler(writer, &slog.HandlerOptions{Level: slog.LevelInfo})
	return slog.New(handler)
}
