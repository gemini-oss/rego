// pkg/common/log/log.go
package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	// Log level constants
	TRACE = iota
	DEBUG
	INFO
	WARNING
	ERROR
	FATAL
	PANIC

	// Color Escape Codes
	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	Reset   = "\033[0m"

	// Formatting
	Bold      = "\033[1m"
	Italic    = "\033[3m"
	Underline = "\033[4m"
)

// String representation for log levels
func LogLevel(level int) string {
	switch level {
	case TRACE:
		return fmt.Sprintf("%sTRACE%s", Blue, Reset)
	case DEBUG:
		return fmt.Sprintf("%sDEBUG%s", Cyan, Reset)
	case INFO:
		return fmt.Sprintf("%sINFO%s", Green, Reset)
	case WARNING:
		return fmt.Sprintf("%sWARNING%s", Yellow, Reset)
	case ERROR:
		return fmt.Sprintf("%sERROR%s", Red, Reset)
	case FATAL:
		return fmt.Sprintf("%sFATAL%s", Magenta, Reset)
	case PANIC:
		return fmt.Sprintf("%sPANIC%s", Magenta, Reset)
	default:
		return fmt.Sprintf("%sUNKNOWN%s", Red, Reset)
	}
}

type Logger struct {
	prefix string         // prefix to write at beginning of each log line
	logger *log.Logger    // standard logger
	out    io.WriteCloser // destination for output
	verbosity int         // log level {TRACE, DEBUG, INFO, WARNING, ERROR, FATAL, PANIC}
}

/*
# log.logf
- logs formatted message at specified level
*/
func (l *Logger) logf(level int, format string, v ...interface{}) {
	l.logger.SetPrefix(l.getPrefix(level))
	l.logger.Printf(format, v...)
}

/*
 * # log.log
 * - logs message at specified level
 */
func (l *Logger) log(level int, v ...interface{}) {
	l.logger.SetPrefix(l.getPrefix(level))
	l.logger.Println(v...)
}

/*
 * # log.Print
 * - logs line at INFO level
 */
func (l *Logger) Print(v ...interface{}) {
	l.log(INFO, v...)
}

/*
 * # log.Printf
 * - logs formatted message at INFO level
 */
func (l *Logger) Printf(format string, v ...interface{}) {
	l.logf(INFO, format, v...)
}

/*
 * # log.Println
 * - logs line at INFO level
 */
func (l *Logger) Println(v ...interface{}) {
	l.log(INFO, v...)
}

/*
 * # log.Trace
 * - logs line at TRACE level
 */
func (l *Logger) Trace(v ...interface{}) {
	l.log(TRACE, v...)
}

/*
 * # log.Tracef
 * - logs formatted message at TRACE level
 */
func (l *Logger) Tracef(format string, v ...interface{}) {
	l.logf(TRACE, format, v...)
}

/*
 * # log.Debug
 * - logs line at DEBUG level
 */
func (l *Logger) Debug(v ...interface{}) {
	l.log(DEBUG, v...)
}

/*
 * # log.Debugf
 * - logs formatted message at DEBUG level
 */
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.logf(DEBUG, format, v...)
}

/*
 * # log.Info
 * - logs line at INFO level
 */
func (l *Logger) Fatal(v ...interface{}) {
	l.log(FATAL, v...)
	os.Exit(1)
}

/*
 * # log.Fatal
 * - logs formatted message at FATAL level and then calls os.Exit(1)
 */
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.logf(FATAL, format, v...)
	os.Exit(1)
}

/*
 * # log.Panic
 * - logs line at PANIC level and then panics
 */
func (l *Logger) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	l.log(PANIC, s)
	panic(s)
}

/*
 * # log.Panicf
 * - logs formatted message at PANIC level and then panics
 */
func (l *Logger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	l.log(PANIC, s)
	panic(s)
}

/*
 * # log.Close
 * - closes the Logger's file if it is a *os.File
 */
func (l *Logger) Close() error {
	if f, ok := l.out.(*os.File); ok {
		return f.Close()
	}
	return nil
}

func (l *Logger) getPrefix(level int) string {
	// Get the caller's filename and line number
	_, path, line, _ := runtime.Caller(3)

	// Get just the filename
	file := filepath.Base(path)

	// Formatted current timestamp
	timestamp := time.Now().Format("2003/12/03 03:05:07")

	timestamp = fmt.Sprintf("%s%s%s", Green, timestamp, Reset)
	file = fmt.Sprintf("%s%s%s", Blue, file, Reset)
	lineColor := fmt.Sprintf("%s%d%s", Yellow, line, Reset)

	/*
		Standard Go Format
		%s %s:%d %s %s ", timestamp, file, line, l.prefix, LogLevel(level))

		Python Format
		[%s] {%s:%d} %s - ", timestamp, file, line, LogLevel(level))
	*/
	return fmt.Sprintf("[%s] {%s:%s} %s - ", timestamp, file, lineColor, LogLevel(level)) // Python Format

}

/*
 * # NewLogger
 * - creates a new Logger with the specified prefix
 */
func NewLogger(prefix string, verbosity int) *Logger {
	LOG_FILE := "./rego.log"
	logFile, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}
	logOut := io.MultiWriter(os.Stdout, logFile)

	logger := log.New(logOut, "", 0)

	// If verbosity is not set, set it to INFO
	if verbosity == 0 {
		verbosity = INFO
	}

	return &Logger{
		prefix: prefix,
		logger: logger,
		out:    logFile,
		verbosity: verbosity,
	}
}
