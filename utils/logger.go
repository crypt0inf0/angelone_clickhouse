package utils

import (
    "context"
    "fmt"
    "net/http"
    "os"
    "path/filepath"
    "time"

    "github.com/google/uuid"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "gopkg.in/natefinch/lumberjack.v2"
)

var (
    Logger *zap.SugaredLogger
)

// Initialize logging system
func InitLogger() error {
    // Configure log rotation
    logRotation := &lumberjack.Logger{
        Filename:   filepath.Join("logs", "app.log"),
        MaxSize:    100,    // megabytes
        MaxAge:     7,      // days
        MaxBackups: 5,
        Compress:   true,   // compress rotated files
        LocalTime:  true,
    }

    // Configure log levels and encoders
    config := zap.NewProductionEncoderConfig()
    config.TimeKey = "timestamp"
    config.EncodeTime = zapcore.ISO8601TimeEncoder
    config.EncodeLevel = zapcore.CapitalLevelEncoder
    config.StacktraceKey = "stacktrace"
    config.CallerKey = "caller"

    // Create JSON encoder
    jsonEncoder := zapcore.NewJSONEncoder(config)

    // Create log level handlers
    highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
        return lvl >= zapcore.ErrorLevel
    })
    lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
        return lvl < zapcore.ErrorLevel
    })

    // Create core with multiple outputs
    core := zapcore.NewTee(
        // Error and above go to error log file
        zapcore.NewCore(jsonEncoder,
            zapcore.AddSync(&lumberjack.Logger{
                Filename:   "logs/error.log",
                MaxSize:    100,
                MaxBackups: 3,
                MaxAge:     7,
                Compress:   true,
            }),
            highPriority,
        ),
        // Info and debug go to main log file
        zapcore.NewCore(jsonEncoder,
            zapcore.AddSync(logRotation),
            lowPriority,
        ),
        // All levels go to console in development
        zapcore.NewCore(jsonEncoder,
            zapcore.AddSync(os.Stdout),
            zapcore.DebugLevel,
        ),
    )

    // Create logger with options
    logger := zap.New(core,
        zap.AddCaller(),
        zap.AddCallerSkip(1),
        zap.AddStacktrace(zapcore.ErrorLevel),
    )

    Logger = logger.Sugar()
    return nil
}

// RequestLogger middleware for HTTP request logging
func RequestLogger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        requestID := uuid.New().String()
        ctx := context.WithValue(r.Context(), "request_id", requestID)

        // Log request
        Logger.Infow("Request started",
            "request_id", requestID,
            "method", r.Method,
            "path", r.URL.Path,
            "remote_addr", r.RemoteAddr,
            "user_agent", r.UserAgent(),
        )

        // Create response wrapper to capture status code
        rw := &responseWriter{w, http.StatusOK}
        
        // Process request
        next.ServeHTTP(rw, r.WithContext(ctx))

        // Log response
        Logger.Infow("Request completed",
            "request_id", requestID,
            "method", r.Method,
            "path", r.URL.Path,
            "status", rw.status,
            "duration_ms", time.Since(start).Milliseconds(),
        )
    })
}

// Error logs an error with stack trace
func Error(err error, msg string, fields ...interface{}) {
    Logger.Errorw(msg,
        append([]interface{}{
            "error", err,
            "stack", fmt.Sprintf("%+v", err),
        }, fields...)...,
    )
}

// Custom response writer to capture status code
type responseWriter struct {
    http.ResponseWriter
    status int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.status = code
    rw.ResponseWriter.WriteHeader(code)
}
