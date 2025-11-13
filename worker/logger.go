package worker

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	maillog "github.com/wneessen/go-mail/log"
)

// Logger is a custom logger wrapper that implements packages' logger interface.
// It delegates all logging operations to zerolog, providing a consistent logging interface.
type Logger struct{}

// NewLogger creates and returns a new Logger instance.
func NewLogger() *Logger {
	return &Logger{}
}

// * asynq logger interface implementation
// Print logs a message at the specified zerolog level.
// This is the core logging method that all other log level methods delegate to.
func (*Logger) Print(level zerolog.Level, args ...interface{}) {
	log.WithLevel(level).Msg(fmt.Sprint(args...))
}

// Debug logs a message at Debug level.
func (logger *Logger) Debug(args ...interface{}) {
	logger.Print(zerolog.DebugLevel, args...)
}

// Info logs a message at Info level.
func (logger *Logger) Info(args ...interface{}) {
	logger.Print(zerolog.InfoLevel, args...)
}

// Warn logs a message at Warning level.
func (logger *Logger) Warn(args ...interface{}) {
	logger.Print(zerolog.WarnLevel, args...)
}

// Error logs a message at Error level.
func (logger *Logger) Error(args ...interface{}) {
	logger.Print(zerolog.ErrorLevel, args...)
}

// Fatal logs a message at Fatal level and the process will exit with status set to 1.
func (logger *Logger) Fatal(args ...interface{}) {
	logger.Print(zerolog.FatalLevel, args...)
}

// * go-redis logger interface implementation

func (*Logger) Printf(ctx context.Context, format string, v ...interface{}) {
	log.WithLevel(zerolog.DebugLevel).Msgf(format, v...)
}

//* go-mail Logger interface implementation

// directionString converts a mail log Direction to a readable string.
// enum that refers to direction of SMTP communication between client and the SMTP server 
func directionString(d maillog.Direction) string {
	if d == maillog.DirClientToServer {
		return "client_to_server"
	}
	return "server_to_client"
}

// Debugf logs a mail debug message.
func (logger *Logger) Debugf(l maillog.Log) {
	log.WithLevel(zerolog.DebugLevel).
		Str("direction", directionString(l.Direction)).
		Msgf(l.Format, l.Messages...)
}

// Infof logs a mail info message.
func (logger *Logger) Infof(l maillog.Log) {
	log.WithLevel(zerolog.InfoLevel).
		Str("direction", directionString(l.Direction)).
		Msgf(l.Format, l.Messages...)
}

// Warnf logs a mail warning message.
func (logger *Logger) Warnf(l maillog.Log) {
	log.WithLevel(zerolog.WarnLevel).
		Str("direction", directionString(l.Direction)).
		Msgf(l.Format, l.Messages...)
}

// Errorf logs a mail error message.
func (logger *Logger) Errorf(l maillog.Log) {
	log.WithLevel(zerolog.ErrorLevel).
		Str("direction", directionString(l.Direction)).
		Msgf(l.Format, l.Messages...)
}
