package parser

type NALUType uint

const (
	NALU_NONIDR NALUType = 1
	NALU_IDR    NALUType = 5
	NALU_SEI    NALUType = 6
	NALU_SPS    NALUType = 7
	NALU_PPS    NALUType = 8
	NALU_AUD    NALUType = 9

	SEI_TYPE_USER_DATA_UNREGISTERED = 5
)

//go:generate enumer -type=NALUType -trimprefix=NALU_

type NALUInfo struct {
	NalRefIdc uint
	Type      NALUType
}
