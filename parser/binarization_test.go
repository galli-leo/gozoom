package parser

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const notOk uint = 0xffffffff

type binCase struct {
	bins   uint
	binIdx uint
	output uint
}

func runBinTest(t *testing.T, b Binarization, cases []binCase) {
	for idx, tCase := range cases {
		t.Run(fmt.Sprintf("case_%d", idx), func(t *testing.T) {
			ok, val := b.GetValue(tCase.bins, tCase.binIdx)
			if tCase.output == notOk {
				assert.Falsef(t, ok, "expected get value to fail!")
			} else {
				assert.Truef(t, ok, "expected get value to succeed!")
				assert.EqualValuesf(t, tCase.output, val, "expected output to match")
			}
		})
	}
}

func assertUnique(t *testing.T, b Binarization, vals map[uint]uint, i uint, bins uint) {
	ok, val := b.GetValue(bins, i)
	if ok {
		if prevVal, k := vals[val]; k {
			t.Errorf("Got duplicated result value %d for bin: 0x%x, previously at bin: 0x%x", val, bins, prevVal)
		} else {
			vals[val] = bins
		}
	}
}

func ensureUnique(t *testing.T, b Binarization, vals map[uint]uint, maxBinIdx uint, i uint, bins uint) {
	if i > maxBinIdx {
		return
	}
	newBins := bins | (0 << i)
	assertUnique(t, b, vals, i, newBins)
	ensureUnique(t, b, vals, maxBinIdx, i+1, newBins)

	newBins = bins | (1 << i)
	assertUnique(t, b, vals, i, newBins)
	ensureUnique(t, b, vals, maxBinIdx, i+1, newBins)
}

func ensureUniqueness(t *testing.T, b Binarization, maxBinIdx uint) {
	t.Run("EnsureUniqueness", func(t *testing.T) {
		ensureUnique(t, b, map[uint]uint{}, maxBinIdx, 0, 0)
	})
}

func TestMBTypeIBin(t *testing.T) {
	runBinTest(t, MbTypeIBin, []binCase{
		{
			0, 0, uint(I_NxN),
		},
		{
			0b000001, 5, uint(I_16x16_0_0_0),
		},
		{
			0b100001, 5, uint(I_16x16_1_0_0),
		},
	})

	// ensureUniqueness(t, MbTypeIBin, 6)
}

func TestFixedLen1Bin(t *testing.T) {
	runBinTest(t, NewFLBin(1), []binCase{
		{
			0, 0, 0,
		},
		{
			1, 0, 1,
		},
		{
			0b10, 1, notOk,
		},
	})
}

func TestUbin(t *testing.T) {
	runBinTest(t, NewUBin(), []binCase{
		{
			0, 0, 0,
		},
		{
			0b0111, 3, 3,
		},
		{
			0b0111, 2, notOk,
		},
	})
}
