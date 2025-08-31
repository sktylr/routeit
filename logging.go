package routeit

import (
	"context"
	"log/slog"
	"os"
)

// The [LogAttrExtractor] can be used to output additional request metadata to
// the default log line that is produced on every request. By default, routeit
// will log the request path (including the rewritten and edge path), the HTTP
// status of the response, the request's User-Agent field, the request's method
// and the client's IP address.
type LogAttrExtractor func(*Request, HttpStatus) []slog.Attr

type logger struct {
	log        *slog.Logger
	extraAttrs LogAttrExtractor
}

func newLogger(handler slog.Handler, debug bool, fn LogAttrExtractor) *logger {
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
	if fn == nil {
		fn = func(r *Request, hs HttpStatus) []slog.Attr { return []slog.Attr{} }
	}
	return &logger{log: slog.New(jsonHandler), extraAttrs: fn}
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

	l.log.LogAttrs(req.Context(), level, "Received request", l.attrs(rw, req)...)

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

func (l *logger) attrs(rw *ResponseWriter, req *Request) []slog.Attr {
	base := []slog.Attr{
		slog.String("method", req.mthd.name),
		slog.String("path", req.Path()),
		slog.String("edge_path", req.uri.EdgePathString()),
		slog.String("raw_path", req.RawPath()),
		slog.Int("status", int(rw.s.code)),
		slog.String("user_agent", req.userAgent),
		slog.String("client_ip", req.ip),
	}
	return append(base, l.extraAttrs(req, rw.s)...)
}
