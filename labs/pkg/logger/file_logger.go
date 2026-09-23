package logger

import (
	"os"
	"path/filepath"
)

type FileLogger struct {
	file string
}

func NewFileLogger(filename string) *FileLogger {

	path := filepath.Join(filename)

	return &FileLogger{
		file: path,
	}
}

func (logger *FileLogger) Log(info string) {
	f, _ := os.OpenFile(logger.file, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	defer f.Close()
	f.WriteString(info)
}

func (logger *FileLogger) Error(err error) {
	f, _ := os.OpenFile(logger.file, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	defer f.Close()
	f.WriteString(err.Error())
}

func (logger *FileLogger) Warn(warn string) {
	f, _ := os.OpenFile(logger.file, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	defer f.Close()
	f.WriteString(warn)
}
