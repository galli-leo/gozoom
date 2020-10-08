package parser

import (
	"github.com/hashicorp/go-multierror"
)

// func newReader(nalu []byte) *NaluReader {
// 	cleaned := h264.RemoveH264orH265EmulationBytes(nalu)
// 	slice := cleaned[1:] // First byte is nalu metadata
// 	buf := bytes.NewBuffer(slice)
// 	return &NaluReader{
// 		r:   &GolombBitReader{R: buf},
// 		log: newLogger("NaluReader"),
// 	}
// }

// type NaluReader struct {
// 	r   *GolombBitReader
// 	log *zap.SugaredLogger
// }

// func (p *NaluReader) ReadUE() uint {
// 	ret, err := p.r.ReadExponentialGolombCode()
// 	if err != nil {
// 		p.log.Errorw("Failed to read unsigned golomb", "error", err)
// 	}
// 	return ret
// }

// func (p *NaluReader) ReadSE() int {
// 	ret, err := p.r.ReadSE()
// 	if err != nil {
// 		p.log.Errorw("Failed to read signed golomb", "error", err)
// 	}
// 	return int(ret)
// }

// func (p *NaluReader) ReadBits(n int) uint {
// 	ret, err := p.r.ReadBits(n)
// 	if err != nil {
// 		p.log.Errorw("Failed to read unsigned", "error", err)
// 		panic("read failed")
// 	}
// 	return ret
// }

// func (p *NaluReader) Align() {
// 	p.r.left = 0
// }

// // TODO: This is not spec compliant!
// func (p *NaluReader) MoreRBSPData() bool {
// 	if p.r.R.Len() == 0 && p.r.left == 0 {
// 		return false
// 	}

// 	return true
// }

func NewGolombReader(r BitReader) *GolombBitReader {
	return &GolombBitReader{BitReader: r}
}

type GolombBitReader struct {
	BitReader
	err error
}

func (r *GolombBitReader) addError(err error) {
	if err != nil {
		r.err = multierror.Append(r.err, err)
	}
}

func (r *GolombBitReader) ReadUE() (res uint) {
	i := uint8(0)
	for {
		bit := r.ReadBits(1)
		if !(bit == 0 && i < 32) {
			break
		}
		i++
	}
	res = r.ReadBits(i)
	res += (1 << uint(i)) - 1
	return
}

func (r *GolombBitReader) ReadSE() (res int) {
	num := r.ReadUE()
	res = int(num)
	if res&0x01 != 0 {
		res = (res + 1) / 2
	} else {
		res = -res / 2
	}
	return
}

func (p *GolombBitReader) MoreRBSPData() bool {
	return true
}

func (r *GolombBitReader) Error() error {
	err := r.BitReader.Error()
	if r.err != nil || err != nil {
		err1 := multierror.Prefix(r.err, "Errors from GolombBitReader")
		err2 := multierror.Prefix(err, "Errors from Underlying BitReader")
		return multierror.Append(err1, err2)
	}

	return nil
}
