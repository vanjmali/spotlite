package logging

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/vanjmali/spotlite/common-lib/utils"
	"go.opentelemetry.io/otel/trace"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	initOnce sync.Once
)

// Init configures process-wide logging.
// Output is written to both stdout and a rotated file.
func Init(serviceName string) error {
	var initErr error

	initOnce.Do(func() {
		logDir := utils.GetEnv("LOG_DIR", "logs")

		if err := os.MkdirAll(logDir, 0o750); err != nil {
			initErr = fmt.Errorf("failed to create log directory: %w", err)
			return
		}

		maxSizeMB := utils.GetPositiveIntEnv("LOG_ROTATE_MAX_SIZE_MB", 20)
		maxBackups := utils.GetPositiveIntEnv("LOG_ROTATE_MAX_BACKUPS", 5)
		maxAgeDays := utils.GetPositiveIntEnv("LOG_ROTATE_MAX_AGE_DAYS", 14)
		compress := utils.GetBoolEnv("LOG_ROTATE_COMPRESS", true)

		logFilePath := filepath.Join(logDir, fmt.Sprintf("%s.log", serviceName))
		if err := ensureLogFilePermissions(logFilePath); err != nil {
			initErr = fmt.Errorf("failed to prepare log file: %w", err)
			return
		}

		fileWriter := &lumberjack.Logger{
			Filename:   logFilePath,
			MaxSize:    maxSizeMB,
			MaxBackups: maxBackups,
			MaxAge:     maxAgeDays,
			Compress:   compress,
		}

		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.LUTC | log.Lmsgprefix)
		log.SetPrefix(fmt.Sprintf("service=%s ", serviceName))
		log.SetOutput(io.MultiWriter(os.Stdout, fileWriter))

		Infof(context.Background(), "logger initialized file=%s rotate_size_mb=%d rotate_backups=%d rotate_age_days=%d",
			logFilePath, maxSizeMB, maxBackups, maxAgeDays)
	})

	return initErr
}

func Infof(ctx context.Context, format string, args ...any) {
	log.Printf("level=INFO "+withContextFields(ctx, format), args...)
}

func Warnf(ctx context.Context, format string, args ...any) {
	log.Printf("level=WARN "+withContextFields(ctx, format), args...)
}

func Errorf(ctx context.Context, format string, args ...any) {
	log.Printf("level=ERROR "+withContextFields(ctx, format), args...)
}

func Securityf(ctx context.Context, format string, args ...any) {
	log.Printf("level=WARN category=security "+withContextFields(ctx, format), args...)
}

func Auditf(ctx context.Context, format string, args ...any) {
	log.Printf("level=INFO category=audit "+withContextFields(ctx, format), args...)
}

func withContextFields(ctx context.Context, format string) string {
	if ctx == nil {
		return format
	}

	if span := trace.SpanFromContext(ctx); span != nil {
		sc := span.SpanContext()
		if !sc.IsValid() {
			return format
		}

		return "trace_id=" + sc.TraceID().String() + " span_id=" + sc.SpanID().String() + " " + format
	}

	return format
}

func ensureLogFilePermissions(path string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE, 0o640)
	if err != nil {
		return err
	}
	_ = f.Close()

	return os.Chmod(path, 0o640)
}
