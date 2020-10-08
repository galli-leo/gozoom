package parser

import (
	"fmt"
	"math"
	"math/bits"
)

func NewUBin() *UnaryBin {
	return &UnaryBin{}
}

type UnaryBin struct {
}

func (b *UnaryBin) GetValue(bins uint, binIdx uint) (ok bool, val uint) {
	lastBin := bins >> binIdx
	if lastBin&1 == 0 {
		ok = uint(bits.OnesCount(bins)) == binIdx
		val = binIdx
	}

	return
}

func NewTUBin(cMax uint) *TruncUnaryBin {
	return &TruncUnaryBin{
		u:    NewUBin(),
		cMax: cMax,
	}
}

type TruncUnaryBin struct {
	u    *UnaryBin
	cMax uint
}

func (b *TruncUnaryBin) GetValue(bins uint, binIdx uint) (ok bool, val uint) {
	if binIdx < b.cMax {
		ok, val = b.u.GetValue(bins, binIdx)
	} else {
		ok = uint(bits.OnesCount(bins)) == binIdx // Truncated
		val = b.cMax
	}
	return
}

var FlagBin = NewFLBin(1)

func NewFLBin(cMax uint) *FixedLenBin {
	return &FixedLenBin{
		cMax: cMax,
	}
}

type FixedLenBin struct {
	cMax uint
}

func (b *FixedLenBin) GetValue(bins uint, binIdx uint) (ok bool, val uint) {
	fixedLen := uint(math.Ceil(math.Log2(float64(b.cMax) + 1)))
	if binIdx == fixedLen-1 {
		ok = true
		val = bins
	}

	return
}

func fixIMbTable(inp map[uint]IMBTypeConst) map[uint]uint {
	ret := map[uint]uint{}

	for key, val := range inp {
		ret[key] = val.MBType()
	}

	return ret
}

var MbTypeIBin = NewTblBin(fixIMbTable(map[uint]IMBTypeConst{
	0:         I_NxN,
	0b100000:  I_16x16_0_0_0,
	0b100001:  I_16x16_1_0_0,
	0b100010:  I_16x16_2_0_0,
	0b100011:  I_16x16_3_0_0,
	0b1001000: I_16x16_0_1_0,
	0b1001001: I_16x16_1_1_0,
	0b1001010: I_16x16_2_1_0,
	0b1001011: I_16x16_3_1_0,
	0b1001100: I_16x16_0_2_0,
	0b1001101: I_16x16_1_2_0,
	0b1001110: I_16x16_2_2_0,
	0b1001111: I_16x16_3_2_0,
	0b101000:  I_16x16_0_0_1,
	0b101001:  I_16x16_1_0_1,
	0b101010:  I_16x16_2_0_1,
	0b101011:  I_16x16_3_0_1,

	0b1011000: I_16x16_0_1_1,
	0b1011001: I_16x16_1_1_1,
	0b1011010: I_16x16_2_1_1,
	0b1011011: I_16x16_3_1_1,
	0b1011100: I_16x16_0_2_1,
	0b1011101: I_16x16_1_2_1,
	0b1011110: I_16x16_2_2_1,
	0b1011111: I_16x16_3_2_1,

	0b11: I_PCM,
}))

func NewTblBin(tbl map[uint]uint) *TableBin {
	return &TableBin{tbl}
}

type TableBin struct {
	tbl map[uint]uint
}

func (b *TableBin) GetValue(bins uint, binIdx uint) (ok bool, val uint) {
	rev := bits.Reverse(bins)
	key := rev >> (64 - (binIdx + 1))
	val, ok = b.tbl[key]
	return
}

var GolombBin = NewUIBin("Golomb")

func NewUIBin(name string) *UnimplBin {
	return &UnimplBin{name}
}

type UnimplBin struct {
	name string
}

func (b *UnimplBin) GetValue(bins uint, binIdx uint) (ok bool, val uint) {
	panic(fmt.Sprintf("Binarization %s is not yet implemented!", b.name))
	return
}
