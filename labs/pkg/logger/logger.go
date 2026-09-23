package logger

type Logger interface {
	Log(info string)
	Error(err error)
	Warn(warn string)
}
