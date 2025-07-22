package routeit

import (
	"context"
	"log/slog"
	"os"
)

type logger struct {
	log *slog.Logger
}

func newLogger(debug bool) *logger {
	logOpts := slog.HandlerOptions{}
	if debug {
		logOpts.Level = slog.LevelDebug
	} else {
		logOpts.Level = slog.LevelInfo
	}
	jsonHandler := slog.NewJSONHandler(os.Stdout, &logOpts)
	return &logger{log: slog.New(jsonHandler)}
}

func (l *logger) LogRequestAndResponse(rw *ResponseWriter, req *Request) {
	var level slog.Level
	switch {
	case rw.s.Is5xx():
		level = slog.LevelError
	case rw.s.Is4xx():
		level = slog.LevelWarn
	default:
		level = slog.LevelInfo
	}

	attrs := []slog.Attr{
		slog.String("method", req.mthd.name),
		slog.String("path", req.Path()),
		slog.String("edge_path", req.uri.edgePath),
		slog.String("raw_path", req.RawPath()),
		slog.Int("status", int(rw.s.code)),
	}

	l.log.LogAttrs(context.Background(), level, "Received request", attrs...)
	// TODO: could dump the request body here if we want to?
}

func (l *logger) Debug(msg string, args ...any) {
	l.log.Debug(msg, args...)
}

func (l *logger) Info(msg string, args ...any) {
	l.log.Info(msg, args...)
}

func (l *logger) Warn(msg string, args ...any) {
	l.log.Warn(msg, args...)
}

func (l *logger) Error(msg string, args ...any) {
	l.log.Error(msg, args...)
}

func (l *logger) Enabled(ctx context.Context, level slog.Level) bool {
	return l.log.Enabled(ctx, level)
}
