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

func Debug(format string, logs ...any) {
	str := fmt.Sprintf(format, logs...)
	logger.Log(LEVEL_DEBUG, str)
}

func Info(format string, logs ...any) {
	str := fmt.Sprintf(format, logs...)
	logger.Log(LEVEL_INFO, str)
}

func Warning(format string, logs ...any) {
	str := fmt.Sprintf(format, logs...)
	logger.Log(LEVEL_WARNING, str)
}

func Error(format string, logs ...any) {
	str := fmt.Sprintf(format, logs...)
	logger.Log(LEVEL_ERROR, str)
}

func Fatal(format string, logs ...any) {
	str := fmt.Sprintf(format, logs...)
	logger.Log(LEVEL_FATAL, str)
}
