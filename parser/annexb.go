package parser

import (
	"fmt"
	"io"

	"github.com/hashicorp/go-multierror"
	"go.uber.org/zap"
)

func NewAnnexBReader(r BitReader, log *zap.SugaredLogger) *AnnexBReader {
	reader := &AnnexBReader{
		BitReader: r,
		log:       log.Named("AnnexBReader"),
	}

	return reader
}

type AnnexBReader struct {
	BitReader
	log  *zap.SugaredLogger
	curr *NALUInfo
	err  error
}

func (r *AnnexBReader) Next() bool {
	numZero := 0
	for {
		b, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				return false
			}
			r.err = multierror.Append(r.err, err)
			return false
		}
		if numZero < 2 {
			if b != 0 {
				r.addErr("Expected zero byte, got 0x%x", b)
				return false
			}
		} else {
			if b != 1 && b != 0 {
				r.addErr("Expected zero/one byte, got 0x%x", b)
				return false
			}
			if b == 1 {
				// parse nalu info
				forbiddenZero := r.ReadBits(1)
				if forbiddenZero != 0 {
					r.addErr("forbidden_zero_bit is not zero")
				}
				r.curr = &NALUInfo{}
				r.curr.NalRefIdc = uint(r.ReadBits(2))
				r.curr.Type = NALUType(r.ReadBits(5))
				return true
			}
		}
		numZero++

	}
	return false
}

func (r *AnnexBReader) Current() *NALUInfo {
	return r.curr
}

func (r *AnnexBReader) Error() error {
	err := r.BitReader.Error()
	if r.err != nil || err != nil {
		err1 := multierror.Prefix(r.err, "AnnexBReader Errors")
		err2 := multierror.Prefix(err, "Errors from Underlying BitReader")
		return multierror.Append(err1, err2)
	}

	return nil
}

func (r *AnnexBReader) addErr(msg string, args ...interface{}) {
	args = append([]interface{}{r.Position(), r.Position()}, args...)
	r.err = multierror.Append(r.err, fmt.Errorf("[%04x/%04d] "+msg, args...))
}
