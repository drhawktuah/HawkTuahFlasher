package core

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type LogLevel uint8

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarning
	LogLevelError
	LogLevelFatal
)

const (
	ansiReset   = "\033[0m"
	ansiGray    = "\033[90m"
	ansiGreen   = "\033[32m"
	ansiYellow  = "\033[33m"
	ansiRed     = "\033[31m"
	ansiMagenta = "\033[35m"
)

type Logger struct {
	output io.Writer
	level  LogLevel
	color  bool
	mutex  sync.Mutex
}

func NewLogger(output io.Writer, level LogLevel) *Logger {
	if output == nil {
		output = os.Stderr
	}

	return &Logger{
		output: output,
		level:  level,
		color:  true,
	}
}

func (logger *Logger) SetLevel(level LogLevel) {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()

	logger.level = level
}

func (logger *Logger) Level() LogLevel {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()

	return logger.level
}

func (logger *Logger) SetColor(enabled bool) {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()

	logger.color = enabled
}

func (logger *Logger) ColorEnabled() bool {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()

	return logger.color
}

func (logger *Logger) Log(level LogLevel, format string, arguments ...any) {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()

	if level < logger.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelName := level.String()

	if logger.color {
		levelName = level.Color() + levelName + ansiReset
	}

	fmt.Fprintf(logger.output, "[%s] [%s] %s\n", timestamp, levelName, fmt.Sprintf(format, arguments...))
}

func (logger *Logger) Debug(format string, arguments ...any) {
	logger.Log(LogLevelDebug, format, arguments...)
}

func (logger *Logger) Info(format string, arguments ...any) {
	logger.Log(LogLevelInfo, format, arguments...)
}

func (logger *Logger) Warning(format string, arguments ...any) {
	logger.Log(LogLevelWarning, format, arguments...)
}

func (logger *Logger) Error(format string, arguments ...any) {
	logger.Log(LogLevelError, format, arguments...)
}

func (logger *Logger) Fatal(format string, arguments ...any) {
	logger.Log(LogLevelFatal, format, arguments...)
}

func (level LogLevel) String() string {
	switch level {
		case LogLevelDebug:
			return "DEBUG"
		case LogLevelInfo:
			return "INFO"
		case LogLevelWarning:
			return "WARNING"
		case LogLevelError:
			return "ERROR"
		case LogLevelFatal:
			return "FATAL"
		default:
			return "UNKNOWN"
	}
}

func (level LogLevel) Color() string {
	switch level {
		case LogLevelDebug:
			return ansiGray
		case LogLevelInfo:
			return ansiGreen
		case LogLevelWarning:
			return ansiYellow
		case LogLevelError:
			return ansiRed
		case LogLevelFatal:
			return ansiMagenta
		default:
			return ""
	}
}

var DefaultLogger = NewLogger(os.Stderr, LogLevelInfo)

func (logger *Logger) LogWriter(writer io.Writer, level LogLevel, format string, arguments ...any) {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()

	if level < logger.level {
		return
	}

	if writer == nil {
		writer = logger.output
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelName := level.String()

	if logger.color {
		levelName = level.Color() + levelName + ansiReset
	}

	fmt.Fprintf(writer, "[%s] [%s] %s\n", timestamp, levelName, fmt.Sprintf(format, arguments...))
}

func Debug(format string, arguments ...any) {
	DefaultLogger.Debug(format, arguments...)
}

func Info(format string, arguments ...any) {
	DefaultLogger.Info(format, arguments...)
}

func Warning(format string, arguments ...any) {
	DefaultLogger.Warning(format, arguments...)
}

func Error(format string, arguments ...any) {
	DefaultLogger.Error(format, arguments...)
}

func Fatal(format string, arguments ...any) {
	DefaultLogger.Fatal(format, arguments...)
}

func Sprintf(format string, arguments ...any) string {
	return fmt.Sprintf(format, arguments...)
}

func Sprint(arguments ...any) string {
	return fmt.Sprint(arguments...)
}

func Sprintln(arguments ...any) string {
	return fmt.Sprintln(arguments...)
}

func Errorf(format string, arguments ...any) error {
	return fmt.Errorf(format, arguments...)
}

func Fprintf(writer io.Writer, format string, arguments ...any) (int, error) {
	return fmt.Fprintf(writer, format, arguments...)
}

func Fprint(writer io.Writer, arguments ...any) (int, error) {
	return fmt.Fprint(writer, arguments...)
}

func Fprintln(writer io.Writer, arguments ...any) (int, error) {
	return fmt.Fprintln(writer, arguments...)
} 

func Printf(writer io.Writer, format string, arguments ...any) {
	fmt.Fprintf(writer, format, arguments...)
}

func Println(writer io.Writer, arguments ...any) {
	fmt.Fprintln(writer, arguments...)
}

func (logger *Logger) Printf(format string, arguments ...any) {
	logger.Log(LogLevelInfo, format, arguments...)
}

func (logger *Logger) Println(arguments ...any) {
	logger.Log(LogLevelInfo, "%s", fmt.Sprintln(arguments...))
}