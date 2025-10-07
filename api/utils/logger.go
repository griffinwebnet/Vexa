package utils

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// LogLevel represents the logging level... captain obvious here but I know some will be surprised...
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// LogRotationConfig defines log rotation settings
type LogRotationConfig struct {
	MaxSize    int64         // Maximum size in bytes before rotation
	MaxFiles   int           // Maximum number of rotated files to keep
	Compress   bool          // Whether to compress rotated files
	RotateTime time.Duration // Time-based rotation interval
}

// AuditEvent represents a detailed audit log entry
type AuditEvent struct {
	Timestamp time.Time              `json:"timestamp"`
	User      string                 `json:"user"`
	Action    string                 `json:"action"`
	Category  string                 `json:"category"`
	Resource  string                 `json:"resource"`
	Details   map[string]interface{} `json:"details"`
	IPAddress string                 `json:"ip_address,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
	Success   bool                   `json:"success"`
	Error     string                 `json:"error,omitempty"`
}

// LogFile represents a single log file with rotation capabilities
type LogFile struct {
	file     *os.File
	logger   *log.Logger
	path     string
	config   LogRotationConfig
	lastSize int64
	lastTime time.Time
	mutex    sync.RWMutex
}

// Logger handles application logging with rotation
type Logger struct {
	debugLog *LogFile
	infoLog  *LogFile
	warnLog  *LogFile
	errorLog *LogFile
	auditLog *LogFile
	vexaLog  *LogFile // Combined log file
	logDir   string
	config   LogRotationConfig
	verbose  bool
}

var (
	// Global logger instance
	globalLogger *Logger
	logMutex     sync.RWMutex
)

// Default rotation config
var defaultConfig = LogRotationConfig{
	MaxSize:    10 * 1024 * 1024, // 10MB
	MaxFiles:   6,                // Keep 6 rotated files
	Compress:   true,             // Compress old files
	RotateTime: 24 * time.Hour,   // Rotate daily
}

// InitLogger initializes the global logger with rotation
func InitLogger() error {
	return InitLoggerWithConfig(defaultConfig, true)
}

// GetLogDirectory returns the appropriate log directory to all functions that need it
func GetLogDirectory() string {
	return "/var/log/vexa"
}

// InitLoggerWithConfig initializes the logger with custom configuration
func InitLoggerWithConfig(config LogRotationConfig, verbose bool) error {
	logMutex.Lock()
	defer logMutex.Unlock()

	logDir := GetLogDirectory()
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	logger := &Logger{
		logDir:  logDir,
		config:  config,
		verbose: verbose,
	}

	// Initialize individual log files...
	// logs issues using fmt.Errorf directly so that a failed initalization
	// will dump the error to stdout/stderr which captures to journalctl
	// this way the service doesnt just fail out silently without
	// logging the failure anywhere.
	var err error
	logger.debugLog, err = newLogFile(filepath.Join(logDir, "debug.log"), config)
	if err != nil {
		return fmt.Errorf("failed to create debug log: %v", err)
	}

	logger.infoLog, err = newLogFile(filepath.Join(logDir, "info.log"), config)
	if err != nil {
		return fmt.Errorf("failed to create info log: %v", err)
	}

	logger.warnLog, err = newLogFile(filepath.Join(logDir, "warn.log"), config)
	if err != nil {
		return fmt.Errorf("failed to create warn log: %v", err)
	}

	logger.errorLog, err = newLogFile(filepath.Join(logDir, "error.log"), config)
	if err != nil {
		return fmt.Errorf("failed to create error log: %v", err)
	}

	logger.auditLog, err = newLogFile(filepath.Join(logDir, "audit.log"), config)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %v", err)
	}

	// Create combined vexa.log file for general log viewing
	logger.vexaLog, err = newLogFile(filepath.Join(logDir, "vexa.log"), config)
	if err != nil {
		return fmt.Errorf("failed to create vexa log: %v", err)
	}

	globalLogger = logger

	// Log startup message to the log files
	logger.Info("Vexa logging system initialized with rotation enabled")
	logger.Info("Log directory: %s", logDir)
	logger.Info("Max file size: %d bytes", config.MaxSize)
	logger.Info("Max rotated files: %d", config.MaxFiles)
	logger.Info("Compression: %v", config.Compress)
	logger.Info("Verbose logging: %v", verbose)

	return nil
}

// newLogFile creates a new LogFile with rotation capabilities
func newLogFile(path string, config LogRotationConfig) (*LogFile, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	// Get file info for initial size
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	logFile := &LogFile{
		file:     file,
		logger:   log.New(file, "", log.LstdFlags),
		path:     path,
		config:   config,
		lastSize: info.Size(),
		lastTime: time.Now(),
	}

	return logFile, nil
}

// write writes a message to the log file with rotation check
func (lf *LogFile) write(level, message string) error {
	lf.mutex.Lock()
	defer lf.mutex.Unlock()

	// Check if rotation is needed
	if lf.needsRotation() {
		if err := lf.rotate(); err != nil {
			return fmt.Errorf("failed to rotate log file: %v", err)
		}
	}

	// Write the message
	timestamp := time.Now().Format("2006/01/02 15:04:05")
	logMessage := fmt.Sprintf("[%s] %s %s", timestamp, level, message)

	_, err := lf.file.WriteString(logMessage + "\n")
	if err != nil {
		return err
	}

	// Update size and time
	if info, err := lf.file.Stat(); err == nil {
		lf.lastSize = info.Size()
		lf.lastTime = time.Now()
	}

	return nil
}

// needsRotation checks if the log file needs rotation
func (lf *LogFile) needsRotation() bool {
	// Check size-based rotation
	if info, err := lf.file.Stat(); err == nil {
		if info.Size() >= lf.config.MaxSize {
			return true
		}
	}

	// Check time-based rotation
	if time.Since(lf.lastTime) >= lf.config.RotateTime {
		return true
	}

	return false
}

// rotate performs log file rotation
func (lf *LogFile) rotate() error {
	// Close current file
	if err := lf.file.Close(); err != nil {
		return err
	}

	// Rotate existing files
	if err := lf.rotateFiles(); err != nil {
		return err
	}

	// Create new log file
	file, err := os.OpenFile(lf.path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	lf.file = file
	lf.logger = log.New(file, "", log.LstdFlags)
	lf.lastSize = 0
	lf.lastTime = time.Now()

	return nil
}

// rotateFiles handles the rotation of numbered log files
func (lf *LogFile) rotateFiles() error {
	basePath := lf.path
	dir := filepath.Dir(basePath)
	baseName := filepath.Base(basePath)

	// Find existing rotated files
	var rotatedFiles []string
	for i := 1; i <= lf.config.MaxFiles; i++ {
		rotatedPath := filepath.Join(dir, fmt.Sprintf("%s.%d", baseName, i))
		if _, err := os.Stat(rotatedPath); err == nil {
			rotatedFiles = append(rotatedFiles, rotatedPath)
		}
	}

	// Sort by number (reverse order)
	sort.Slice(rotatedFiles, func(i, j int) bool {
		numI := extractNumber(rotatedFiles[i])
		numJ := extractNumber(rotatedFiles[j])
		return numI > numJ
	})

	// Move files up the chain
	for i := len(rotatedFiles) - 1; i >= 0; i-- {
		currentNum := extractNumber(rotatedFiles[i])
		if currentNum >= lf.config.MaxFiles {
			// Delete the oldest file
			os.Remove(rotatedFiles[i])
			continue
		}

		newPath := filepath.Join(dir, fmt.Sprintf("%s.%d", baseName, currentNum+1))
		newPath = strings.TrimSuffix(newPath, ".tar.gz")

		// Move the file
		if err := os.Rename(rotatedFiles[i], newPath); err != nil {
			return err
		}

		// Compress if needed and it's not the most recent
		if lf.config.Compress && currentNum > 1 {
			if err := lf.compressFile(newPath); err != nil {
				// Log error but don't fail rotation
				fmt.Printf("Warning: Failed to compress %s: %v\n", newPath, err)
			}
		}
	}

	// Move current log to .1
	newPath := filepath.Join(dir, fmt.Sprintf("%s.1", baseName))
	return os.Rename(basePath, newPath)
}

// extractNumber extracts the rotation number from a file path
func extractNumber(path string) int {
	baseName := filepath.Base(path)
	parts := strings.Split(baseName, ".")
	if len(parts) < 2 {
		return 0
	}

	// Handle .tar.gz extension
	if len(parts) >= 3 && parts[len(parts)-2] == "tar" && parts[len(parts)-1] == "gz" {
		if num, err := strconv.Atoi(parts[len(parts)-3]); err == nil {
			return num
		}
	}

	// Handle regular .N extension
	if num, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
		return num
	}

	return 0
}

// compressFile compresses a log file using tar.gz
func (lf *LogFile) compressFile(filePath string) error {
	// Open the file to compress
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create compressed file
	compressedPath := filePath + ".tar.gz"
	compressedFile, err := os.Create(compressedPath)
	if err != nil {
		return err
	}
	defer compressedFile.Close()

	// Create gzip writer
	gzipWriter := gzip.NewWriter(compressedFile)
	defer gzipWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// Get file info
	info, err := file.Stat()
	if err != nil {
		return err
	}

	// Create tar header
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = filepath.Base(filePath)

	// Write header
	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}

	// Copy file content
	_, err = io.Copy(tarWriter, file)
	if err != nil {
		return err
	}

	// Remove original file
	return os.Remove(filePath)
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	logMutex.RLock()
	defer logMutex.RUnlock()

	if globalLogger == nil {
		// Fallback to stdout if not initialized
		return &Logger{
			verbose: true,
		}
	}
	return globalLogger
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if l.verbose {
		if l.debugLog != nil {
			l.debugLog.write("DEBUG", message)
		}
		if l.vexaLog != nil {
			l.vexaLog.write("DEBUG", message)
		}
	}
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if l.infoLog != nil {
		l.infoLog.write("INFO", message)
	}
	if l.vexaLog != nil {
		l.vexaLog.write("INFO", message)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if l.warnLog != nil {
		l.warnLog.write("WARN", message)
	}
	if l.vexaLog != nil {
		l.vexaLog.write("WARN", message)
	}
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if l.errorLog != nil {
		l.errorLog.write("ERROR", message)
	}
	if l.vexaLog != nil {
		l.vexaLog.write("ERROR", message)
	}
}

// Audit logs a detailed audit event
func (l *Logger) Audit(event AuditEvent) {
	// Ensure timestamp is set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(event)
	if err != nil {
		l.Error("Failed to marshal audit event: %v", err)
		return
	}

	// Write to audit log
	if l.auditLog != nil {
		l.auditLog.write("AUDIT", string(jsonData))
	}

	// Also write to combined log
	if l.vexaLog != nil {
		l.vexaLog.write("AUDIT", string(jsonData))
	}
}

// Close closes all log files
func (l *Logger) Close() error {
	var err error

	if l.debugLog != nil && l.debugLog.file != nil {
		if closeErr := l.debugLog.file.Close(); closeErr != nil {
			err = closeErr
		}
	}
	if l.infoLog != nil && l.infoLog.file != nil {
		if closeErr := l.infoLog.file.Close(); closeErr != nil {
			err = closeErr
		}
	}
	if l.warnLog != nil && l.warnLog.file != nil {
		if closeErr := l.warnLog.file.Close(); closeErr != nil {
			err = closeErr
		}
	}
	if l.errorLog != nil && l.errorLog.file != nil {
		if closeErr := l.errorLog.file.Close(); closeErr != nil {
			err = closeErr
		}
	}
	if l.auditLog != nil && l.auditLog.file != nil {
		if closeErr := l.auditLog.file.Close(); closeErr != nil {
			err = closeErr
		}
	}
	if l.vexaLog != nil && l.vexaLog.file != nil {
		if closeErr := l.vexaLog.file.Close(); closeErr != nil {
			err = closeErr
		}
	}

	return err
}

// Convenience functions for global logger
func Debug(format string, args ...interface{}) {
	GetLogger().Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	GetLogger().Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	GetLogger().Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	GetLogger().Error(format, args...)
}

func Audit(event AuditEvent) {
	GetLogger().Audit(event)
}
