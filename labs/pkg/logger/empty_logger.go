package logger

type EmptyLogger struct {
}

func NewEmptyLogger() *EmptyLogger {

	return &EmptyLogger{}
}

func (logger *EmptyLogger) Log(info string) {
}

func (logger *EmptyLogger) Error(err error) {
}

func (logger *EmptyLogger) Warn(warn string) {
}
