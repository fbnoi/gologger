package gologger

import (
	"context"
	"fmt"
	"sync"
)

var (
	logger = &Logger{
		drivers: make(map[Driver]*Formatter),
		mu:      sync.RWMutex{},
	}
)

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
	Debugc(context.Background(), fmt.Sprintf(format, logs...))
}

func Debugc(ctx context.Context, format string, logs ...any) {
	str := fmt.Sprintf(format, logs...)
	logger.Log(ctx, LEVEL_DEBUG, str)
}

func Info(format string, logs ...any) {
	Infoc(context.Background(), format, logs...)
}

func Infoc(ctx context.Context, format string, logs ...any) {
	str := fmt.Sprintf(format, logs...)
	logger.Log(ctx, LEVEL_INFO, str)
}

func Warning(format string, logs ...any) {
	Warningc(context.Background(), format, logs...)
}

func Warningc(ctx context.Context, format string, logs ...any) {
	str := fmt.Sprintf(format, logs...)
	logger.Log(ctx, LEVEL_WARNING, str)
}

func Error(format string, logs ...any) {
	Errorc(context.Background(), format, logs...)
}

func Errorc(ctx context.Context, format string, logs ...any) {
	str := fmt.Sprintf(format, logs...)
	logger.Log(ctx, LEVEL_ERROR, str)
}

func Fatal(format string, logs ...any) {
	Fatalc(context.Background(), format, logs...)
}

func Fatalc(ctx context.Context, format string, logs ...any) {
	str := fmt.Sprintf(format, logs...)
	logger.Log(ctx, LEVEL_FATAL, str)
}
