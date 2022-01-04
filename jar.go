package jt

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var _ io.Closer = (*JarFile)(nil)

type JarFile struct {
	archive *zip.Reader
	io.Closer
}

type readerAtCloser interface {
	io.ReaderAt
	io.Closer
}

func OpenJarFile(name string) (*JarFile, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	stat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat: %w", err)
	}
	return NewJarFile(f, stat.Size())
}

func NewJarFile(rd readerAtCloser, size int64) (*JarFile, error) {
	archive, err := zip.NewReader(rd, size)
	if err != nil {
		return nil, fmt.Errorf("open zip: %w", err)
	}

	return &JarFile{
		archive: archive,
		Closer:  rd,
	}, nil
}

func (f *JarFile) OpenClass(name string) (*Class, error) {
	classFile, err := f.archive.Open(name + ".class")
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	class, err := ParseClass(classFile)
	if err != nil {
		return nil, fmt.Errorf("parse class: %w", err)
	}

	return class, nil
}

func (f *JarFile) ListClasses() []string {
	res := make([]string, 0)

	for _, file := range f.archive.File {
		if filepath.Ext(file.Name) == ".class" {
			res = append(res, strings.TrimSuffix(file.Name, ".class"))
		}
	}

	return res
}
