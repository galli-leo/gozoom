package parser

import (
	"io"

	"go.uber.org/zap"
)

func NewEmuPreventionReader(r BitReader, log *zap.SugaredLogger) *EmuPreventionReader {
	reader := &EmuPreventionReader{
		r:      r,
		log:    log.Named("EmuPreventionReader"),
		bufPos: 3,
	}
	reader.BitReader = NewBitioReader(reader)

	return reader
}

// Reads NAL Unit data and removes the emulation prevention bytes
type EmuPreventionReader struct {
	// reader using this as an io.Reader
	BitReader
	// underlying reader
	r      BitReader
	log    *zap.SugaredLogger
	buf    [3]byte
	bufPos byte
}

func (r *EmuPreventionReader) Reset() {
	r.bufPos = 3
}

func (r *EmuPreventionReader) ReadByte() (b byte, err error) {
	if r.bufPos == 3 {
		r.bufPos = 0
		n, err := r.r.Read(r.buf[:])
		if err != nil && err != io.EOF {
			return 0, err
		}
		if n == 0 {
			return 0, io.EOF
		}
		if r.buf == [3]byte{0x00, 0x00, 0x03} {
			r.bufPos++
			r.buf = [3]byte{0x00, 0x00, 0x00}
		}
	}

	val := r.buf[r.bufPos]
	r.bufPos++
	return val, nil
}

func (r *EmuPreventionReader) Read(b []byte) (n int, err error) {
	var val byte
	for n = 0; n < len(b); n++ {
		val, err = r.ReadByte()
		if err != nil {
			return
		}
		b[n] = val
	}

	return
}
