package errorlog

import (
	"log"
	"os"
)

type Logger struct {
	file *os.File
	*log.Logger
}

// New creates a new error logger that writes to the specified log file.
func New(logFile string) (*Logger, error) {
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &Logger{
		file,
		log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}, nil

}

// WriteError writes an error message to the log file.
func (l *Logger) WriteError(err error) {
	errStr := err.Error()
	l.Println(errStr)
}

// WriteInfo writes an informational message to the log file.
// The prefix is temporarily changed to "INFO: " for this message.
func (l *Logger) WriteInfo(info string) {
	l.SetPrefix("INFO: ")
	l.Println(info)
	l.SetPrefix("ERROR: ")
}

// Close closes the log file.
func (l *Logger) Close() {
	l.file.Close()
}
