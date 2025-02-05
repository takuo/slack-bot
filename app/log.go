package app

import (
	"context"
	"log/slog"
)

type logger struct {
	log    *slog.Logger
	prefix string
}

func (l *logger) Output(_ int, str string) error {
	l.log.Log(context.Background(), slog.LevelDebug, l.prefix+str)
	return nil
}
