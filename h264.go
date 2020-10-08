package main

// #cgo CFLAGS: -g -Wall
// #cgo pkg-config: openh264
// #include "h264.h"
import "C"
import (
	"unsafe"
)

var num = 0

func DecodeFrame(buf []byte) []byte {
	size := 3440 * 1440
	// chroma := make([]C.uchar, size)
	// cb := make([]C.uchar, size)
	// cr := make([]C.uchar, size)
	output := make([]byte, 3*size)
	C.DecodeOneFrame((*C.uchar)(unsafe.Pointer(&buf[0])), C.int(len(buf)), (**C.uchar)(unsafe.Pointer(&output[0])))
	return output
}
