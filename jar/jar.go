package jar

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tsatke/jt/class"
)

var _ io.Closer = (*File)(nil)

type File struct {
	archive *zip.Reader
	io.Closer
}

type readerAtCloser interface {
	io.ReaderAt
	io.Closer
}

func Open(name string) (*File, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	stat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat: %w", err)
	}
	return New(f, stat.Size())
}

func New(rd readerAtCloser, size int64) (*File, error) {
	archive, err := zip.NewReader(rd, size)
	if err != nil {
		return nil, fmt.Errorf("open zip: %w", err)
	}

	return &File{
		archive: archive,
		Closer:  rd,
	}, nil
}

func (f *File) OpenClass(name string) (*class.Class, error) {
	classFile, err := f.archive.Open(name + ".class")
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	class, err := class.ParseClass(classFile)
	if err != nil {
		return nil, fmt.Errorf("parse class: %w", err)
	}

	return class, nil
}

func (f *File) ListClasses() []string {
	res := make([]string, 0)

	for _, file := range f.archive.File {
		if filepath.Ext(file.Name) == ".class" {
			res = append(res, strings.TrimSuffix(file.Name, ".class"))
		}
	}

	return res
}
