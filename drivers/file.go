package drivers

import (
	"sync"

	"fbnoi.com/gologger"
)

func NewFileDriver(format string) *FileDriver {
	return &FileDriver{
		writers: make(map[gologger.Level]*FileWriter),
		format:  format,
		mu:      &sync.RWMutex{},
	}
}

type FileDriver struct {
	writers map[gologger.Level]*FileWriter
	format  string
	mu      *sync.RWMutex
}

func (fd *FileDriver) AddWriter(level gologger.Level, w *FileWriter) {
	fd.mu.Lock()
	defer fd.mu.Unlock()

	fd.writers[level] = w
}

func (fd *FileDriver) Log(level gologger.Level, log string) {
	fd.mu.RLock()
	defer fd.mu.RUnlock()

	if writer, ok := fd.writers[level]; ok {
		writer.Write([]byte(log))
	}
}

func (fd *FileDriver) SetFormat(format string) {
	fd.format = format
}

func (fd *FileDriver) GetFormat() string {
	return fd.format
}

func (fd *FileDriver) Close() (err error) {
	for _, w := range fd.writers {
		err = w.Close()
	}

	return
}
