package classfile

import "io"

var _ OffsetReader = (*offsetReader)(nil)

type OffsetReader interface {
	io.Reader
	Offset() uint
}

func NewOffsetReader(rd io.Reader) OffsetReader {
	return &offsetReader{
		Reader: rd,
		offset: 0,
	}
}

type offsetReader struct {
	io.Reader
	offset uint
}

func (o *offsetReader) Offset() uint {
	return o.offset
}

func (o *offsetReader) Read(p []byte) (n int, err error) {
	n, err = o.Reader.Read(p)
	o.offset += uint(n)
	return
}
