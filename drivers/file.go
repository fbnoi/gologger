package drivers

import (
	"fbnoi/gologger"
	"sync"
)

type FileDriver struct {
	writers []*FileWriter
	async   bool
	wg      *sync.WaitGroup
	format  string
}

func (fd *FileDriver) Log(level gologger.Level, log string) {
	if fd.async {
		for _, w := range fd.writers {
			fd.wg.Add(1)
			go func(w *FileWriter) {
				w.Write([]byte(log))
				fd.wg.Done()
			}(w)
		}
		fd.wg.Wait()
	} else {
		for _, w := range fd.writers {
			w.Write([]byte(log))
		}
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
