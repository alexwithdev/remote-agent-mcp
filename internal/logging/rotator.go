package logging

import "gopkg.in/natefinch/lumberjack.v2"

// newRotator returns a lumberjack writer that rotates the log file when it
// exceeds maxSizeMB, retaining at most maxBackups compressed backups.
func newRotator(filename string, maxSizeMB, maxBackups int) *lumberjack.Logger {
	if maxSizeMB <= 0 {
		maxSizeMB = 10
	}
	if maxBackups <= 0 {
		maxBackups = 5
	}
	return &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSizeMB,
		MaxBackups: maxBackups,
		Compress:   true,
	}
}
