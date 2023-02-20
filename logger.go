package gologger

import (
	"fmt"
	"sync"
)

var (
	LEVEL_DEBUG   Level = 00000001
	LEVEL_INFO    Level = 00000010
	LEVEL_WARNING Level = 00000100
	LEVEL_ERROR   Level = 00001000
	LEVEL_FATAL   Level = 00010000

	LEVEL_NAME = map[Level]string{
		LEVEL_DEBUG:   "Debug",
		LEVEL_INFO:    "Info",
		LEVEL_WARNING: "Warning",
		LEVEL_ERROR:   "Error",
		LEVEL_FATAL:   "Fatal",
	}
)

type Level int

func (l Level) GetName() string {
	return LEVEL_NAME[l]
}

type Driver interface {
	Log(Level, string)
	SetFormat(string)
	GetFormat() string
	Close() error
}

type Logger struct {
	drivers map[Driver]*Formatter
	mu      sync.RWMutex
}

func (l *Logger) Log(lv Level, a ...any) {
	ps := map[string]any{"L": lv.GetName(), "M": fmt.Sprintf("%v ", a...)}
	l.mu.RLock()
	defer l.mu.RUnlock()

	for d, f := range l.drivers {
		d.Log(lv, f.Format(ps))
	}
}

func (l *Logger) AddDriver(d Driver) {
	format := d.GetFormat()
	formatter := newFormatter(format)
	l.mu.Lock()
	defer l.mu.Unlock()

	l.drivers[d] = formatter
}
