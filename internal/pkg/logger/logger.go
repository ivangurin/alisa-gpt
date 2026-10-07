package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"alisa-gpt/internal/pkg/metadata"

	"go.opentelemetry.io/otel/trace"
)

const (
	LogFormatDev  string = "dev"
	LogFormatText string = "text"
)

const (
	fieldTraceID   string = "trace_id"
	fieldSpanID    string = "span_id"
	FieldRequestID string = "request_id"
)

type Logger interface {
	With(key, value string) Logger
	Debug(ctx context.Context, message string, fields ...any)
	Info(ctx context.Context, message string, fields ...any)
	Warn(ctx context.Context, message string, fields ...any)
	Error(ctx context.Context, message string, fields ...any)
}

// logger оборачивает *slog.Logger и предоставляет map-ориентированный API: Info("msg", map[string]any{...})
type logger struct {
	sl *slog.Logger
}

// NewLogger создаёт логгер на базе slog.
//
//	logLevel  — debug | info | warn | error (по умолчанию info)
//	logFormat:
//	  dev   — цветной человекочитаемый вывод для локальной разработки
//	  text  — однострочный текст без цветов
//	  json  — структурированный JSON (по умолчанию)
//
// extraHandlers, если переданы, получают каждую запись лога в дополнение к
// основному handler.
func NewLogger(out io.Writer, logLevel, logFormat string, extraHandlers ...slog.Handler) Logger {
	var level slog.Level
	if logLevel != "" {
		_ = level.UnmarshalText([]byte(logLevel))
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	switch strings.ToLower(logFormat) {
	case LogFormatDev:
		handler = NewPrettyHandler(out, opts)
	case LogFormatText:
		handler = slog.NewTextHandler(out, opts)
	default:
		handler = slog.NewJSONHandler(out, opts)
	}

	if len(extraHandlers) > 0 {
		handlers := append([]slog.Handler{handler}, extraHandlers...)
		handler = slog.NewMultiHandler(handlers...)
	}

	return &logger{sl: slog.New(handler)}
}

func (l *logger) With(key, value string) Logger {
	return &logger{
		sl: l.sl.With(slog.String(key, value)),
	}
}

func (l *logger) Debug(ctx context.Context, message string, values ...any) {
	l.log(ctx, slog.LevelDebug, message, values)
}

func (l *logger) Info(ctx context.Context, message string, values ...any) {
	l.log(ctx, slog.LevelInfo, message, values)
}

func (l *logger) Warn(ctx context.Context, message string, values ...any) {
	l.log(ctx, slog.LevelWarn, message, values)
}

func (l *logger) Error(ctx context.Context, message string, values ...any) {
	l.log(ctx, slog.LevelError, message, values)
}

func (l *logger) log(ctx context.Context, level slog.Level, message string, values []any) {
	var attrs []slog.Attr
	if len(values) == 1 {
		if m, ok := values[0].(map[string]any); ok {
			attrs = make([]slog.Attr, 0, len(m))
			for k, v := range m {
				attrs = append(attrs, slog.Any(k, v))
			}
		}
	} else if len(values) > 1 {
		attrs = make([]slog.Attr, 0, len(values)/2)
		for i := 0; i < len(values)-1; i += 2 {
			k, ok := values[i].(string)
			if !ok {
				k = fmt.Sprint(values[i])
			}
			attrs = append(attrs, slog.Any(k, values[i+1]))
		}
	}

	fields := l.getFields(ctx)
	attrs = append(attrs, fields...)

	l.sl.LogAttrs(ctx, level, message, attrs...)
}

func (l *logger) getFields(ctx context.Context) []slog.Attr {
	res := make([]slog.Attr, 0, 3)

	traceID := l.getTraceID(ctx)
	if traceID != "" {
		res = append(res, slog.Any(fieldTraceID, traceID))
	}

	spanID := l.getSpanID(ctx)
	if spanID != "" {
		res = append(res, slog.Any(fieldSpanID, spanID))
	}

	requestID := l.getRequestID(ctx)
	if requestID != "" {
		res = append(res, slog.Any(FieldRequestID, requestID))
	}

	return res
}

func (l *logger) getTraceID(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		return spanCtx.TraceID().String()
	}

	return ""
}

func (l *logger) getSpanID(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasSpanID() {
		return spanCtx.SpanID().String()
	}

	return ""
}

func (l *logger) getRequestID(ctx context.Context) string {
	return metadata.GetRequestID(ctx)
}
