package parser

import (
	"io"

	"github.com/hashicorp/go-multierror"
	"github.com/icza/bitio"
)

type BitReader interface {
	io.Reader
	io.ByteReader
	// Align aligns the bit stream to a byte boundary, so next read will read/use data from the next byte. Returns the number of unread / skipped bits.
	Align() (skipped uint8)

	// ReadBits reads n bits and returns them as the lowest n bits of u.
	ReadBits(n uint8) (u uint)

	// Returns the error encountered while reading, if any.
	Error() error

	Position() uint
}

func NewBitioReader(r io.Reader) BitReader {
	return &BitioReader{Reader: bitio.NewReader(r)}
}

type BitioReader struct {
	*bitio.Reader
	pos uint
	err error
}

func (b *BitioReader) Error() error {
	return b.err
}

func (b *BitioReader) Read(p []byte) (n int, err error) {
	n, err = b.Reader.Read(p)
	b.pos += uint(n)
	return
}

func (b *BitioReader) ReadBits(n uint8) (u uint) {
	if n == 0 {
		return
	}
	val, err := b.Reader.ReadBits(n)
	u = uint(val)
	if err != nil {
		b.err = multierror.Append(b.err, err)
	}

	b.pos += uint(n)

	return
}

func (b *BitioReader) Position() uint {
	return b.pos
}
