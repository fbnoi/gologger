package gologger

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	formats = map[string]func(map[string]any) string{
		"T": longTime,
		"t": shortTime,
		"D": longDate,
		"d": shortDate,
		"L": keyFactory("L"),
		"M": keyFactory("M"),
	}

	internalReg = regexp.MustCompile("%[a-zA-Z]")
)

func newFormatter(format string) *Formatter {
	f := &Formatter{
		builderPool: sync.Pool{New: func() any {
			return &strings.Builder{}
		}},
	}
	poss := internalReg.FindAllStringIndex(format, -1)
	var (
		length          = len(format)
		cursor          = 0
		key, str, label string
	)
	for _, pos := range poss {
		str = format[cursor:pos[0]]
		label = format[pos[0]:pos[1]]
		f.funcs = append(f.funcs, textFactory(str))
		key = strings.Trim(label, "%")
		if fn, ok := formats[key]; ok {
			f.funcs = append(f.funcs, fn)
		} else {
			f.funcs = append(f.funcs, textFactory(label))
		}
		cursor = pos[1]
	}
	if cursor < length {
		f.funcs = append(f.funcs, textFactory(format[cursor:length]))
	}

	return f
}

type Formatter struct {
	funcs       []func(map[string]any) string
	builderPool sync.Pool
}

func (f *Formatter) Format(ms map[string]any) string {
	sb := f.builderPool.Get().(*strings.Builder)
	defer func() {
		sb.Reset()
		f.builderPool.Put(sb)
	}()
	for _, fn := range f.funcs {
		sb.WriteString(fn(ms))
	}

	return sb.String()
}

func textFactory(t string) func(map[string]any) string {
	return func(map[string]any) string {
		return t
	}
}

func keyFactory(key string) func(map[string]any) string {
	return func(ms map[string]any) string {
		if m, ok := ms[key]; ok {
			if str, ok := m.(string); ok {
				return str
			}

			return fmt.Sprint(m)
		}

		return ""
	}
}

// time format 15:04:05.000
func longTime(map[string]any) string {
	return time.Now().Format("15:04:05.000")
}

// time format 15:04:05.000
func shortTime(map[string]any) string {
	return time.Now().Format("15:04")
}

// time format 2006-01-02
func longDate(map[string]any) string {
	return time.Now().Format("2006-01-02")
}

// time format 01-02
func shortDate(map[string]any) string {
	return time.Now().Format("01-02")
}
