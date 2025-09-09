// pkg/common/log/log.go
package log

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"
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

// ansiRegex matches ANSI color escape sequences
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// stripANSI removes ANSI color codes from a string
func stripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// colorStripWriter wraps an io.Writer and strips ANSI color codes
type colorStripWriter struct {
	w io.Writer
}

func (csw *colorStripWriter) Write(p []byte) (n int, err error) {
	stripped := stripANSI(string(p))
	return csw.w.Write([]byte(stripped))
}

// rotatingFile manages log file rotation
type rotatingFile struct {
	mu          sync.Mutex
	filename    string
	file        *os.File
	maxSize     int64 // Maximum size in bytes before rotation
	maxBackups  int   // Maximum number of backup files to keep
	currentSize int64
	compress    bool // Whether to compress rotated files
}

// newRotatingFile creates a new rotating file writer
func newRotatingFile(filename string, maxSize int64, maxBackups int, compress bool) (*rotatingFile, error) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	return &rotatingFile{
		filename:    filename,
		file:        file,
		maxSize:     maxSize,
		maxBackups:  maxBackups,
		currentSize: info.Size(),
		compress:    compress,
	}, nil
}

func (rfw *rotatingFile) Write(p []byte) (n int, err error) {
	rfw.mu.Lock()
	defer rfw.mu.Unlock()

	// Check if rotation is needed
	if rfw.currentSize+int64(len(p)) > rfw.maxSize {
		if err := rfw.rotate(); err != nil {
			return 0, err
		}
	}

	n, err = rfw.file.Write(p)
	rfw.currentSize += int64(n)
	return n, err
}

func (rfw *rotatingFile) rotate() error {
	// Close current file
	if err := rfw.file.Close(); err != nil {
		return err
	}

	// Determine file extension based on compression
	ext := ""
	if rfw.compress {
		ext = ".gz"
	}

	// Clean up oldest backup if we're at max
	oldestBackup := fmt.Sprintf("%s.%d%s", rfw.filename, rfw.maxBackups, ext)
	os.Remove(oldestBackup) // Ignore error if file doesn't exist

	// Rename existing files
	for i := rfw.maxBackups - 1; i > 0; i-- {
		oldName := fmt.Sprintf("%s.%d%s", rfw.filename, i, ext)
		newName := fmt.Sprintf("%s.%d%s", rfw.filename, i+1, ext)
		os.Rename(oldName, newName) // Ignore errors for non-existent files
	}

	// Move current file to .1 and optionally compress
	rotatedName := fmt.Sprintf("%s.1", rfw.filename)
	if err := os.Rename(rfw.filename, rotatedName); err != nil {
		return err
	}

	// Compress the rotated file if needed
	if rfw.compress {
		if err := compressFile(rotatedName, rotatedName+".gz"); err != nil {
			// Log error but don't fail rotation
			log.Printf("Failed to compress rotated log: %v", err)
		} else {
			// Remove uncompressed file after successful compression
			os.Remove(rotatedName)
		}
	}

	// Create new file
	file, err := os.OpenFile(rfw.filename, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	rfw.file = file
	rfw.currentSize = 0
	return nil
}

// compressFile compresses a file using gzip
func compressFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	gzipWriter := gzip.NewWriter(dstFile)
	defer gzipWriter.Close()

	_, err = io.Copy(gzipWriter, srcFile)
	return err
}

func (rfw *rotatingFile) Close() error {
	rfw.mu.Lock()
	defer rfw.mu.Unlock()
	return rfw.file.Close()
}

/*
 * @param {int} level - log level
 * @param {bool} color - enable/disable colorized output
 * @return {string} - string representation of log level
 */
func LogLevel(level int, color bool) string {
	lvl := ""
	switch level {
	case TRACE:
		lvl = "TRACE"
	case DEBUG:
		lvl = "DEBUG"
	case INFO:
		lvl = "INFO"
	case WARNING:
		lvl = "WARNING"
	case ERROR:
		lvl = "ERROR"
	case FATAL:
		lvl = "FATAL"
	case PANIC:
		lvl = "PANIC"
	default:
		lvl = "UNKNOWN"
	}

	if color {
		color := getColor(level)
		return fmt.Sprintf("%s%s%s", color, lvl, Reset)
	}

	return lvl
}

func getColor(level int) string {
	switch level {
	case TRACE:
		return Blue
	case DEBUG:
		return Cyan
	case INFO:
		return Green
	case WARNING:
		return Yellow
	case ERROR:
		return Red
	case FATAL:
		return Magenta
	case PANIC:
		return Magenta
	default:
		return Red
	}
}

type Logger struct {
	Color     bool           // enable/disable colorized output
	prefix    string         // prefix to write at beginning of each log line
	logger    *log.Logger    // standard logger
	out       io.WriteCloser // destination for output
	Verbosity int            // log level {TRACE, DEBUG, INFO, WARNING, ERROR, FATAL, PANIC}
	fileOnly  bool           // if true, only log to file (no stdout)
}

/*
# log.logf
- logs formatted message at specified level
*/
func (l *Logger) logf(level int, format string, v ...interface{}) {
	if level >= l.Verbosity {
		l.logger.SetPrefix(l.getPrefix(level))
		l.logger.Printf(format, v...)
	}
}

/*
 * # log.log
 * - logs message at specified level
 */
func (l *Logger) log(level int, v ...interface{}) {
	if level >= l.Verbosity {
		l.logger.SetPrefix(l.getPrefix(level))
		l.logger.Println(v...)
	}
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
 * # log.Info
 * - logs line at INFO level
 */
func (l *Logger) Info(v ...interface{}) {
	l.log(INFO, v...)
}

/*
 * # log.Infof
 * - logs formatted message at INFO level
 */
func (l *Logger) Infof(format string, v ...interface{}) {
	l.logf(INFO, format, v...)
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
 * # log.Warning
 * - logs formatted message at WARNING level
 */
func (l *Logger) Warning(v ...interface{}) {
	l.log(WARNING, v...)
}

/*
 * # log.Warningf
 * - logs formatted message at WARNING level
 */
func (l *Logger) Warningf(format string, v ...interface{}) {
	l.logf(WARNING, format, v...)
}

/*
 * # log.Error
 * - logs formatted message at ERROR level
 */
func (l *Logger) Error(v ...interface{}) {
	l.log(ERROR, v...)
}

/*
 * # log.Errorf
 * - logs formatted message at ERROR level
 */
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.logf(ERROR, format, v...)
}

/*
 * # log.Fatal
 * - logs line at FATAL level
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

/*
 * # log.Delete
 * - Delete the Logger's file if it is a *os.File
 */
func (l *Logger) Delete() error {
	if f, ok := l.out.(*os.File); ok {
		f.Close()
		return os.Remove(f.Name())
	}
	return nil
}

func (l *Logger) getPrefix(level int) string {
	// Get the caller's filename and line number
	_, path, line, _ := runtime.Caller(3)

	// Get just the filename
	file := filepath.Base(path)

	// Formatted current timestamp
	timestamp := time.Now().Format("2006/01/02 03:04:05 PM")

	var lineColor string
	switch l.Color {
	case true:
		timestamp = fmt.Sprintf("%s%s%s", Green, timestamp, Reset)
		file = fmt.Sprintf("%s%s%s", Blue, file, Reset)
		lineColor = fmt.Sprintf("%s%d%s", Yellow, line, Reset)
	default:
		lineColor = fmt.Sprintf("%d", line)
	}

	/*
		Standard Go Format
		%s %s:%d %s %s ", timestamp, file, line, l.prefix, LogLevel(level))

		Python Format
		[%s] {%s:%d} %s - ", timestamp, file, line, LogLevel(level))
	*/
	return fmt.Sprintf("[%s] %s {%s:%s} %s - ", timestamp, l.prefix, file, lineColor, LogLevel(level, l.Color))
}

// SetOutput sets the output destination for the logger.
func (l *Logger) SetOutput(output io.Writer) {
	l.logger.SetOutput(output)
}

// SetNewFile sets the output destination for the logger to a new file.
func (l *Logger) SetNewFile(logFilePath string) {
	LOG_FILE := logFilePath
	logFile, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}
	// Create a color stripping writer for file output
	fileWriter := &colorStripWriter{w: logFile}
	logOut := io.MultiWriter(os.Stdout, fileWriter)
	l.logger = log.New(logOut, "", 0)
	l.out = logFile
}

// LogOption is a function that modifies a Logger
type LogOption func(*Logger)

// WithStdout configures the logger to only output to stdout
func WithStdout() LogOption {
	return func(l *Logger) {
		l.logger.SetOutput(os.Stdout)
		l.out = nopCloser{os.Stdout}
	}
}

// WithLogFile configures the logger to output to a specific file and stdout
func WithLogFile(filepath string) LogOption {
	return func(l *Logger) {
		logFile, err := os.OpenFile(filepath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			log.Panic(err)
		}
		// Create a color stripping writer for file output
		fileWriter := &colorStripWriter{w: logFile}
		logOut := io.MultiWriter(os.Stdout, fileWriter)
		l.logger.SetOutput(logOut)
		l.out = logFile
	}
}

// WithRotatingFile configures the logger with a rotating log file
func WithRotatingFile(filepath string, maxSizeMB int64, maxBackups int, compress bool) LogOption {
	return func(l *Logger) {
		rotatingFile, err := newRotatingFile(filepath, maxSizeMB*1024*1024, maxBackups, compress)
		if err != nil {
			log.Panic(err)
		}
		// Create a color stripping writer for file output
		fileWriter := &colorStripWriter{w: rotatingFile}
		logOut := io.MultiWriter(os.Stdout, fileWriter)
		l.logger.SetOutput(logOut)
		l.out = rotatingFile
	}
}

// WithFile configures the logger to only output to file (no stdout)
func WithFile(filepath string) LogOption {
	return func(l *Logger) {
		logFile, err := os.OpenFile(filepath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			log.Panic(err)
		}
		// Always strip colors when writing to file only
		fileWriter := &colorStripWriter{w: logFile}
		l.logger.SetOutput(fileWriter)
		l.out = logFile
		l.fileOnly = true
	}
}

// WithColor enables/disables colorized output
func WithColor(enabled bool) LogOption {
	return func(l *Logger) {
		l.Color = enabled
	}
}

// nopCloser wraps an io.Writer to implement io.WriteCloser with a no-op Close method
type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

/*
 * # NewLogger
 * - creates a new Logger with the specified prefix
 */
func NewLogger(prefix string, verbosity int, opts ...LogOption) *Logger {
	// If verbosity is not set, set it to INFO
	if verbosity == 0 {
		verbosity = INFO
	}

	// Create logger with default settings
	logger := &Logger{
		Color:     true,
		prefix:    prefix,
		logger:    log.New(os.Stdout, "", 0),
		out:       nopCloser{os.Stdout},
		Verbosity: verbosity,
	}

	// If no options provided, use default behavior with rotation
	if len(opts) == 0 {
		LOG_FILE := "./rego.log"
		// Use rotating file writer with:
		// - 100MB max file size
		// - Keep 5 backup files
		// - Enable compression
		rotatingFile, err := newRotatingFile(LOG_FILE, 100*1024*1024, 5, true)
		if err != nil {
			log.Panic(err)
		}
		// Create a color stripping writer for file output
		fileWriter := &colorStripWriter{w: rotatingFile}
		logOut := io.MultiWriter(os.Stdout, fileWriter)
		logger.logger.SetOutput(logOut)
		logger.out = rotatingFile
	} else {
		// Apply provided options
		for _, opt := range opts {
			opt(logger)
		}
	}

	return logger
}
