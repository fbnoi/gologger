package gologger

import (
	"bytes"
	"io"
	"sync"
	"time"

	"github.com/pkg/errors"
)

var (
	write_time_out       time.Duration
	buffer_max_size      int
	buffer_tube_max_size int
)

var lg *logger

type logger struct {
	outs    []io.Writer
	mu      sync.RWMutex
	prefix  string
	format  string
	bufPool sync.Pool
	tube    chan *bytes.Buffer
}

func AddOut(out io.Writer) {
	lg.outs = append(lg.outs, out)
}

func SetPrefix(prefix string) {
	lg.mu.Lock()
	defer lg.mu.Unlock()

	lg.prefix = prefix
}

func SetFormat(format string) {
	lg.mu.Lock()
	defer lg.mu.Unlock()

	lg.format = format
}

func write(p []byte) (n int, err error) {
	var (
		start   = 0
		cursor  = start
		length  = len(p)
		endFlag = make(chan int)
		buffer  *bytes.Buffer
	)
	go func() {
		var _n int
		for cursor < length {
			cursor = cursor + buffer_max_size
			if cursor > length {
				cursor = length
			}
			buffer = getBuffer()

			if _n, err = buffer.Write(p[start:cursor]); err != nil {
				n = n + _n
				endFlag <- 1
			}
			lg.tube <- buffer
		}
		endFlag <- 0
	}()
	timeout := time.NewTimer(write_time_out * time.Millisecond)
	select {
	case <-timeout.C:
		err = errors.Errorf("log channel is full, discard log")
		return
	case <-endFlag:
		return
	}
}

func getBuffer() *bytes.Buffer {
	return lg.bufPool.Get().(*bytes.Buffer)
}

func releaseBuffer(buff *bytes.Buffer) {
	buff.Reset()
	lg.bufPool.Put(buff)
}
