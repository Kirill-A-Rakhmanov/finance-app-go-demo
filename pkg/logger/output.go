package logger

import (
	"io"

	"gopkg.in/natefinch/lumberjack.v2"
)

type FileConfig struct {
	Path       string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
}

func NewOutput(w io.Writer, cfg FileConfig) (io.Writer, func() error) {
	if cfg.Path == "" {
		return w, func() error { return nil }
	}

	file := &lumberjack.Logger{
		Filename:   cfg.Path,
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAgeDays,
	}

	return io.MultiWriter(w, file), file.Close
}
