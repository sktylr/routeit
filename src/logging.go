package routeit

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type logger struct {
	log *slog.Logger
}

func newLogger(handler slog.Handler, debug bool) *logger {
	if handler != nil {
		return &logger{log: slog.New(handler)}
	}
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
		slog.String("edge_path", "/"+strings.Join(req.uri.edgePath, "/")),
		slog.String("raw_path", req.RawPath()),
		slog.Int("status", int(rw.s.code)),
		slog.String("user_agent", req.userAgent),
		slog.String("client_ip", req.ip),
	}

	l.log.LogAttrs(context.Background(), level, "Received request", attrs...)

	if rw.s.isError() && len(req.body) != 0 {
		l.log.Debug("Request failed", slog.String("body", string(req.body)))
	}
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
