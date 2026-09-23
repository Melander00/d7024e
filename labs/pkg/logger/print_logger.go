package logger

import "fmt"

type PrintLogger struct {
	prefix string
}

func NewPrintLogger(prefix string) *PrintLogger {
	return &PrintLogger{
		prefix: prefix,
	}
}

func (logger *PrintLogger) Log(info string) {
	fmt.Printf("[INFO] %s: %s", logger.prefix, info)
}

func (logger *PrintLogger) Error(err error) {
	fmt.Printf("[ERROR] %s: %s", logger.prefix, err)
}

func (logger *PrintLogger) Warn(warn string) {
	fmt.Printf("[WARN] %s: %s", logger.prefix, warn)
}
