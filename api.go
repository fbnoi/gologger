package gologger

import (
	"fmt"
	"sync"
)

var (
	logger = &Logger{
		drivers: make(map[Driver]*Formatter),
		mu:      sync.RWMutex{},
	}
)

func NewLogger() *Logger {
	return &Logger{
		mu: sync.RWMutex{},
	}
}

func init() {
	logger.AddDriver(NewStdout())
}

func AddDriver(d Driver) {
	logger.AddDriver(d)
}

func SetLogger(l *Logger) {
	logger = l
}

func Debug(logs ...any) {
	str := fmt.Sprint(logs...)
	logger.Log(LEVEL_DEBUG, str)
}

func Debugf(format string, logs ...any) {
	Debug(fmt.Sprintf(format, logs...))
}

func Info(logs ...any) {
	str := fmt.Sprint(logs...)
	logger.Log(LEVEL_INFO, str)
}

func Infof(format string, logs ...any) {
	Info(fmt.Sprintf(format, logs...))
}

func Warning(logs ...any) {
	str := fmt.Sprint(logs...)
	logger.Log(LEVEL_WARNING, str)
}

func Warningf(format string, logs ...any) {
	Warning(fmt.Sprintf(format, logs...))
}

func Error(logs ...any) {
	str := fmt.Sprint(logs...)
	logger.Log(LEVEL_ERROR, str)
}

func Errorf(format string, logs ...any) {
	Error(fmt.Sprintf(format, logs...))
}

func Fatal(logs ...any) {
	str := fmt.Sprint(logs...)
	logger.Log(LEVEL_FATAL, str)
}

func Fatalf(format string, logs ...any) {
	Fatal(fmt.Sprintf(format, logs...))
}

func Close() {
	for d := range logger.drivers {
		d.Close()
	}
}
