package classfile

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

type contentReader struct {
	OffsetReader
	byteOrder binary.ByteOrder
}

func newContentReader(rd io.Reader, byteOrder binary.ByteOrder) *contentReader {
	if or, ok := rd.(OffsetReader); ok {
		return &contentReader{
			OffsetReader: or,
			byteOrder:    byteOrder,
		}
	}

	return &contentReader{
		OffsetReader: NewOffsetReader(rd),
		byteOrder:    byteOrder,
	}
}

func (rd *contentReader) uint8() uint8 {
	return rd.raw(1)[0]
}

func (rd *contentReader) uint16() uint16 {
	return rd.byteOrder.Uint16(rd.raw(2))
}

func (rd *contentReader) uint32() uint32 {
	return rd.byteOrder.Uint32(rd.raw(4))
}

func (rd *contentReader) uint64() uint64 {
	return rd.byteOrder.Uint64(rd.raw(8))
}

func (rd *contentReader) float32() float32 {
	return math.Float32frombits(rd.uint32())
}

func (rd *contentReader) float64() float64 {
	return math.Float64frombits(rd.uint64())
}

func (rd *contentReader) raw(n uint) []byte {
	b := make([]byte, n)

	read, err := rd.Read(b)
	if err == io.EOF {
		if uint(read) != n {
			panic(fmt.Errorf("want to read %d, but only read %d, then EOF", n, read))
		}
	} else if err != nil {
		panic(err)
	}

	return b
}
