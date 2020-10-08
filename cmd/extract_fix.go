package cmd

import (
	"bytes"
	"fmt"

	"github.com/icza/bitio"
)

var SequenceToFix = []byte{0x00, 0x00, 0x00, 0x01, 0x67, 0x64, 0x01, 0x32}
var Replacement = []byte{0x00, 0x00, 0x00, 0x01, 0x67, 0x64, 0x04, 0x32}
var StartOfSequence = []byte{0x00, 0x00, 0x00, 0x01, 0x67, 0x64}
var EndOfSequence = []byte{0x01, 0x32}

func TransformFix(b []byte) []byte {
	ret := new(bytes.Buffer)

	w := bitio.NewWriter(ret)
	for idx := bytes.Index(b, SequenceToFix); idx != -1; idx = bytes.Index(b, SequenceToFix) {
		fmt.Printf("Found start: %d\n", idx)
		before := b[:idx]
		w.Write(before)
		w.Write(Replacement)
		// w.Write(StartOfSequence)
		// w.WriteBits(0, 1) // Fix
		// w.Write(EndOfSequence)

		b = b[idx+len(SequenceToFix):]
	}
	w.Write(b)
	// w.Align()

	return ret.Bytes()
}
