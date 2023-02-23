package drivers

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
)

var (
	max_buf_size = 1024 * 3

	default_option = option{
		RotateFormat: "2006-01-02",
		MaxSize:      1024 * 1024 * 3,
		ChanSize:     1024 * 8,
		WriteTimeout: time.Second,
	}
)

func NewFileWriter(dir, fname string, fns ...Option) (*FileWriter, error) {
	opt := default_option
	for _, fn := range fns {
		fn(&opt)
	}
	fi, err := os.Stat(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Errorf("read dir %s failed, error: %s", dir, err)
		} else if err = os.MkdirAll(dir, 0755); err != nil {
			return nil, errors.Errorf("create dir %s failed, error: %s", dir, err)
		}
	} else if !fi.IsDir() {
		return nil, errors.Errorf("path %s already exist and is not a dir", dir)
	}
	snapshots, err := parseRotateFiles(dir, fname, opt.RotateFormat)
	if err != nil {
		return nil, errors.Errorf("parse rotate files failed, error: %s", err)
	}
	var path string
	if len(snapshots) > 0 {
		snapshot := snapshots[len(snapshots)-1]
		path = filepath.Join(dir, fmt.Sprintf("%s.%s.%03d", fname, snapshot.fmtName, snapshot.index))
	} else {
		t := time.Now()
		fmtName := t.Format(opt.RotateFormat)
		path = filepath.Join(dir, fmt.Sprintf("%s.%s.%03d", fname, fmtName, 0))
		snapshots = append(snapshots, &rotateSnapshot{fmtName: fmtName, index: 0, t: t})
	}
	wf, err := newWrappedFile(path)
	if err != nil {
		return nil, errors.Errorf("creat wrapped file %s failed, error: %s", path, err)
	}
	stdlog := log.New(os.Stderr, "flog ", log.LstdFlags)
	fw := &FileWriter{
		opt:    opt,
		dir:    dir,
		fname:  fname,
		stdlog: stdlog,
		bufCh:  make(chan *bytes.Buffer, opt.ChanSize),
		pool: &sync.Pool{
			New: func() any {
				return &bytes.Buffer{}
			},
		},
		closed: 0,
		wg:     sync.WaitGroup{},

		wf:    wf,
		files: snapshots,
	}
	fw.wg.Add(1)
	go fw.do()

	return fw, nil
}

func newWrappedFile(path string) (*wrappedFile, error) {
	fh, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	fi, err := fh.Stat()
	if err != nil {
		return nil, err
	}

	return &wrappedFile{fh: fh, fSize: fi.Size()}, nil
}

func parseRotateFiles(dir, fname, rotateFormat string) ([]*rotateSnapshot, error) {
	fs, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var (
		name       string
		pos, index int
		t          time.Time
		files      []*rotateSnapshot
	)
	for _, fi := range fs {
		if !fi.IsDir() && strings.HasPrefix(fi.Name(), fname) {
			name = strings.TrimLeft(fi.Name(), fmt.Sprintf("%s.", fname))
			pos = strings.LastIndex(name, ".")
			if t, err = time.Parse(rotateFormat, name[:pos]); err != nil {
				continue
			}
			if index, err = strconv.Atoi(strings.TrimLeft(name[pos:], ".")); err != nil {
				continue
			}
			files = append(files, &rotateSnapshot{fmtName: name[:pos], t: t, index: index})
		}
	}
	sort.Slice(files, func(i, j int) bool {
		if files[i].t == files[j].t {
			return files[i].index < files[j].index
		}

		return files[i].t.Before(files[j].t)
	})

	return files, nil
}

type wrappedFile struct {
	fh    *os.File
	fSize int64
}

func (wf *wrappedFile) write(bs []byte) (n int, err error) {
	n, err = wf.fh.Write(bs)
	wf.fSize += int64(n)

	return
}

func (wf *wrappedFile) size() int64 {
	return wf.fSize
}

func (wf *wrappedFile) close() error {
	return wf.fh.Close()
}

type rotateSnapshot struct {
	fmtName string
	index   int
	t       time.Time
}

type FileWriter struct {
	opt    option
	dir    string
	fname  string
	stdlog *log.Logger
	bufCh  chan *bytes.Buffer
	pool   *sync.Pool
	closed int32
	wg     sync.WaitGroup

	wf    *wrappedFile
	files []*rotateSnapshot
}

func (fw *FileWriter) Write(bs []byte) (int, error) {
	if atomic.LoadInt32(&fw.closed) == 1 {
		fw.stdlog.Printf("%s\n", bs)

		return 0, errors.New("file writer all ready closed")
	}
	var (
		start   = 0
		cursor  = start
		n       = cursor
		wFlagCh = make(chan int)
	)
	defer close(wFlagCh)

	go func() {
		length := len(bs)
		for cursor < length {
			cursor = cursor + max_buf_size
			if cursor >= length {
				cursor = length
			}
			buf := fw.getBuf()
			buf.Write(bs[start:cursor])
			fw.bufCh <- buf
			start = cursor
			n = cursor
		}
		wFlagCh <- 1
	}()

	if fw.opt.WriteTimeout == 0 {
		select {
		case <-wFlagCh:
			return len(bs), nil
		default:
			return n, errors.New("channel is full, discard log")
		}
	}

	timeout := time.NewTimer(fw.opt.WriteTimeout)
	select {
	case <-timeout.C:
		return n, errors.New("write timeout, discard log")
	case <-wFlagCh:
		return len(bs), nil
	}
}

func (fw *FileWriter) Close() error {
	atomic.StoreInt32(&fw.closed, 1)
	close(fw.bufCh)
	fw.wg.Wait()

	return nil
}

func (fw *FileWriter) do() {
	var (
		err error
		buf *bytes.Buffer
		ok  bool
	)
	for {
		buf, ok = <-fw.bufCh
		fw.rotate(time.Now())
		if ok {
			if err = fw.write(buf.Bytes()); err != nil {
				fw.stdlog.Println(err)
			}
			fw.putBuf(buf)
		}
		if atomic.LoadInt32(&fw.closed) != 1 {
			continue
		}
		break
	}
	for buf = range fw.bufCh {
		if err = fw.write(buf.Bytes()); err != nil {
			fw.stdlog.Println(err)
		}
		fw.putBuf(buf)
	}
	fw.wg.Done()
}

func (fw *FileWriter) write(p []byte) error {
	if fw.wf == nil {
		fw.stdlog.Println("can't write log to file, please check stderr log for detail")
		fw.stdlog.Printf("%s", p)
	}
	_, err := fw.wf.write(p)

	return err
}

func (fw *FileWriter) rotate(t time.Time) {
	var (
		path, realname, fmtName string
		err                     error
		snapshot                *rotateSnapshot
	)
	if fw.opt.MaxFile != 0 && len(fw.files) > fw.opt.MaxFile {
		counter := len(fw.files) - fw.opt.MaxFile
		for i := 0; i < counter; i++ {
			snapshot = fw.files[0]
			realname = fmt.Sprintf("%s.%s.%03d", fw.fname, snapshot.fmtName, snapshot.index)
			path = filepath.Join(fw.dir, realname)
			if err = os.Remove(path); err != nil {
				fw.stdlog.Printf("remove file %s failed, err: %s", path, err)
			}
			fw.files = fw.files[1:]
		}
	}
	fmtName = t.Format(fw.opt.RotateFormat)
	snapshot = fw.files[len(fw.files)-1]
	if snapshot.fmtName != fmtName || (fw.opt.MaxSize > 0 && fw.opt.MaxSize <= fw.wf.size()) {
		if err = fw.wf.close(); err != nil {
			fw.stdlog.Printf("close log file %s failed, err: %s", path, err)
		}
		index := snapshot.index + 1
		if snapshot.fmtName != fmtName {
			index = 0
		}
		realname = fmt.Sprintf("%s.%s.%03d", fw.fname, fmtName, index)
		path = filepath.Join(fw.dir, realname)
		if fw.wf, err = newWrappedFile(path); err != nil {
			fw.stdlog.Printf("create log file %s failed, err: %s", path, err)

			return
		}
		fw.files = append(fw.files, &rotateSnapshot{fmtName: fmtName, index: index, t: t})
	}
}

func (fw *FileWriter) getBuf() *bytes.Buffer {
	return fw.pool.Get().(*bytes.Buffer)
}

func (fw *FileWriter) putBuf(buf *bytes.Buffer) {
	if buf == nil {
		return
	}
	buf.Reset()
	fw.pool.Put(buf)
}

type option struct {
	RotateFormat string
	MaxFile      int
	MaxSize      int64
	ChanSize     int

	WriteTimeout time.Duration
}

type Option func(opt *option)

func RotateFormat(fmt string) Option {
	return func(opt *option) {
		opt.RotateFormat = fmt
	}
}

func MaxFile(n int) Option {
	return func(opt *option) {
		opt.MaxFile = n
	}
}

func MaxSize(n int64) Option {
	return func(opt *option) {
		opt.MaxSize = n
	}
}

func ChanSize(size int) Option {
	return func(opt *option) {
		opt.ChanSize = size
	}
}

func WriteTimeout(timeout time.Duration) Option {
	return func(opt *option) {
		opt.WriteTimeout = timeout
	}
}
