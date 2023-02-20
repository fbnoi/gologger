package drivers

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var logDir = "./var/log"

func touch(dir, name string) {
	os.MkdirAll(dir, 0755)
	fh, err := os.OpenFile(filepath.Join(dir, name), os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	fh.Close()
}

func TestMain(m *testing.M) {
	ret := m.Run()
	time.Sleep(time.Second)
	os.RemoveAll(logDir)
	os.Exit(ret)
}

func TestParseRotateFiles(t *testing.T) {
	dir := filepath.Join(logDir, "test-parse-rotate")
	tm := time.Now()
	names := []string{
		"info.log.2023-01-20.000",
		"info.log.2022-12-20.001",
		"info.log.2023-02-20.002",
		"info.log." + tm.Format("2006-01-02") + ".005",
	}
	for _, name := range names {
		touch(dir, name)
	}
	l, err := parseRotateFiles(dir, "info.log", "2006-01-02")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(l), len(names))
	item := l[len(l)-1]
	assert.Equal(t, item.index, 5)
	assert.Equal(t, item.t.Format("2006-01-02"), tm.Format("2006-01-02"))
}

func TestRotateExist(t *testing.T) {
	dir := filepath.Join(logDir, "test-parse-rotate")
	names := []string{"info.log." + time.Now().Format("2006-01-02") + ".004"}
	for _, name := range names {
		touch(dir, name)
	}
	fw, err := NewFileWriter(
		dir, "info.log",
		WriteTimeout(time.Second),
	)
	if err != nil {
		t.Error(err)
	}
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i)
	}
	for i := 0; i < 10; i++ {
		for i := 0; i < 1024; i++ {
			if _, err = fw.Write(data); err != nil {
				t.Error(err)
			}
		}
		time.Sleep(time.Millisecond * 10)
	}

	fw.Close()
	fis, err := os.ReadDir(dir)
	if err != nil {
		t.Error(err)
	}
	assert.True(t, len(fis) > 3, "expect more than 3 file get %d", len(fis))
}

func TestMaxFile(t *testing.T) {
	fw, err := NewFileWriter(
		logDir+"/test-max-file",
		"info.log",
		MaxSize(1024*1024),
		MaxFile(1),
	)
	if err != nil {
		t.Fatal(err)
	}
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i)
	}
	for i := 0; i < 10; i++ {
		for i := 0; i < 1024; i++ {
			_, err = fw.Write(data)
			if err != nil {
				t.Error(err)
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	fw.Close()
	fis, err := os.ReadDir(logDir + "/test-max-file")
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, len(fis) <= 2, fmt.Sprintf("expect 2 file get %d", len(fis)))
}

func TestMaxFile2(t *testing.T) {
	files := []string{
		"info.log.2022-12-01.000",
		"info.log.2022-12-02.000",
		"info.log.2022-12-03.000",
		"info.log.2022-12-04.000",
		"info.log.2022-12-05.000",
		"info.log.2022-12-05.001",
	}
	for _, file := range files {
		touch(logDir+"/test-max-file2", file)
	}
	fw, err := NewFileWriter(logDir+"/test-max-file2",
		"info.log",
		MaxSize(1024*1024),
		MaxFile(4),
	)
	if err != nil {
		t.Fatal(err)
	}
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i)
	}
	for i := 0; i < 10; i++ {
		for i := 0; i < 1024; i++ {
			_, err = fw.Write(data)
			if err != nil {
				t.Error(err)
			}
		}
	}
	fw.Close()
	fis, err := os.ReadDir(logDir + "/test-max-file2")
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, len(fis) == 4, fmt.Sprintf("expect 4 file get %d", len(fis)))
}

func TestFileWriter(t *testing.T) {
	fw, err := NewFileWriter(logDir+"/testlog", "info.log")
	if err != nil {
		t.Fatal(err)
	}
	defer fw.Close()
	_, err = fw.Write([]byte("Hello World!\n"))
	if err != nil {
		t.Error(err)
	}
}

func BenchmarkFileWriter(b *testing.B) {
	fw, err := NewFileWriter(
		logDir+"/testlog/bench",
		"info.log",
		MaxSize(1024*1024*2), /*8MB*/
	)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		_, err = fw.Write([]byte("Hello World!\n"))
		if err != nil {
			b.Error(err)
		}
	}
	fw.Close()
}
