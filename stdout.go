package gologger

import (
	"io"
	"os"
)

const defaultPattern = "[%D %t][%L]: %M"

func NewStdout() *Stdout {
	return &Stdout{
		out:    os.Stderr,
		format: defaultPattern,
	}
}

type Stdout struct {
	out    io.Writer
	format string
}

func (s *Stdout) Log(lv Level, log string) {
	s.out.Write([]byte(log))
}

func (s *Stdout) SetFormat(format string) {
	s.format = format
}

func (s *Stdout) GetFormat() string {
	return s.format
}

func (s *Stdout) Close() error {
	return nil
}
