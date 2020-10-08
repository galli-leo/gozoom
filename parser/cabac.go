package parser

import (
	"math"

	"go.uber.org/zap"
)

func NewCabacParser(p *H264Parser) *CabacParser {
	return &CabacParser{
		GolombBitReader: p.GolombBitReader,
		log:             p.log.Named("CabacParser"),
		h264:            p,
		s:               p.slice,
		h:               p.slice.h,
		contexts:        [WELS_CONTEXT_COUNT]CabacElement{},
	}
}

type CabacElement struct {
	uiState uint
	uiMPS   uint
}

type CabacParser struct {
	*GolombBitReader
	log      *zap.SugaredLogger
	h264     *H264Parser
	h        *SliceHeader
	s        *SliceParser
	contexts [WELS_CONTEXT_COUNT]CabacElement

	codIOffset uint16
	codIRange  uint16

	bins uint
}

func Clip3(x, y, z int64) int64 {
	if z < x {
		return x
	}
	if z > y {
		return y
	}

	return z
}

func preCtxState(m, n, sliceQP int64) int64 {
	return Clip3(1, 126, ((m*Clip3(0, 51, sliceQP))>>4)+n)
}

func (p *CabacParser) Initialize() {
	sliceQP := 26 + p.h.PPS.PicInitQPMinus26 + p.h.SliceQPDelta
	for i := 0; i < WELS_CONTEXT_COUNT; i++ {
		m := g_kiCabacGlobalContextIdx[i][0][0]
		n := g_kiCabacGlobalContextIdx[i][0][1]
		iPreCtxState := preCtxState(int64(m), int64(n), int64(sliceQP))
		var elem CabacElement
		if iPreCtxState <= 63 {
			elem.uiState = uint(63 - iPreCtxState)
			elem.uiMPS = 0
		} else {
			elem.uiState = uint(iPreCtxState - 64)
			elem.uiMPS = 1
		}
		p.contexts[i] = elem
	}
}

func (p *CabacParser) InitializeDecodeEngine() {
	p.codIRange = WELS_CABAC_HALF
	p.codIOffset = uint16(p.ReadBits(9))
}

type Binarization interface {
	GetValue(bins uint, binIdx uint) (ok bool, value uint)
}

type CabacSE interface {
	SetParser(p *CabacParser)
	Binarization() Binarization
	MaxBinIdx() uint
	GetCtxIdx(binIdx uint) uint
	Bypass() bool
}

func (p *CabacParser) ParseMBType() uint {
	return p.ParseElement(MBTypeSE)
}

func (p *CabacParser) ParseTransformSize8x8Flag() uint {
	return p.ParseElement(TransformSizeFlagSE)
}

func (p *CabacParser) ParseCodedBlockPattern(ChromaArrayType uint) uint {
	luma := uint(p.ParseElement(CodedBlockPatternLumaSE))
	chroma := uint(0)
	if ChromaArrayType != 0 && ChromaArrayType != 3 {
		chroma = uint(p.ParseElement(CodedBlockPatternChromaSE))
	}
	return luma + chroma*16
}

func (p *CabacParser) ParseMbQpDelta() int {
	val := p.ParseElement(MBQPDeltaSE)
	ret := int(math.Ceil(float64(val) / 2))
	if val%2 == 0 {
		ret = -ret
	}
	return ret
}

func (p *CabacParser) ParseIntraChromaPredMode() uint {
	return p.ParseElement(IntraChromaPredModeSE)
}

func (p *CabacParser) ParsePrevIntraPredModeFlag() uint {
	return p.ParseElement(PrevIntraPredModeFlagSE)
}

func (p *CabacParser) ParseRemIntraPredMode() uint {
	return p.ParseElement(RemIntraPredModeSE)
}

func (p *CabacParser) ParseMbFieldDecodingFlag() uint {
	return p.ParseElement(MBFieldDecodingFlagSE)
}

func (p *CabacParser) ParseEndOfSliceFlag() uint {
	return p.ParseElement(EndOfSliceSE)
}

func (p *CabacParser) ParseCodedBlockFlag(block *ResidualBlock) uint {
	CodedBlockFlagSE.block = block
	return p.ParseElement(CodedBlockFlagSE)
}

func (p *CabacParser) ParseSignificantCoeffFlag(block *ResidualBlock) uint {
	SignificantCoeffFlagSE.block = block
	return p.ParseElement(SignificantCoeffFlagSE)
}

func (p *CabacParser) ParseLastSignificantCoeffFlag(block *ResidualBlock) uint {
	LastSignificantCoeffFlagSE.block = block
	return p.ParseElement(LastSignificantCoeffFlagSE)
}

func (p *CabacParser) ParseElement(elem CabacSE) uint {
	p.bins = uint(0)
	elem.SetParser(p)
	for binIdx := uint(0); ; binIdx++ {
		clampedBinIdx := binIdx
		if binIdx > elem.MaxBinIdx() {
			clampedBinIdx = elem.MaxBinIdx()
		}
		ctxIdx := elem.GetCtxIdx(clampedBinIdx)
		bypass := uint(0)
		if elem.Bypass() {
			bypass = 1
		}

		binVal := p.DecodeBin(ctxIdx, bypass)
		p.bins |= binVal << binIdx
		ok, val := elem.Binarization().GetValue(p.bins, binIdx)
		if ok {
			return val
		}
	}
}

func (p *CabacParser) getBin(binIdx uint) uint {
	return (p.bins >> binIdx) & 1
}

const (
	TERMINATE_CTX = 276
)

/* Table 9-39 */
func (p *CabacParser) GetCtxIdxInc(binIdx uint, ctxIdxOffset uint) uint {
	if ctxIdxOffset == 0 || ctxIdxOffset == 3 {
		if binIdx == 0 {
			return p.CtxIdxIncMbType(ctxIdxOffset)
		}
		if binIdx == 1 {
			return TERMINATE_CTX - ctxIdxOffset
		}
		if binIdx == 2 || binIdx == 3 {
			return binIdx + 1
		}
		if binIdx >= 6 {
			return 7
		}
		return p.CtxIdxIncPrior(ctxIdxOffset, binIdx)
	}

	if ctxIdxOffset == 11 || ctxIdxOffset == 24 {
		return p.CtxIdxIncMbSkipFlag()
	}

	if ctxIdxOffset == 14 {
		if binIdx < 2 {
			return binIdx
		}

		return p.CtxIdxIncPrior(ctxIdxOffset, binIdx)
	}

	if ctxIdxOffset == 17 || ctxIdxOffset == 32 {
		if binIdx == 0 {
			return 0
		}
		if binIdx == 1 {
			return TERMINATE_CTX - ctxIdxOffset
		}
		if binIdx < 4 {
			return binIdx - 1
		}
		if binIdx == 4 {
			return p.CtxIdxIncPrior(ctxIdxOffset, binIdx)
		}
		return 3
	}

	if ctxIdxOffset == 21 {
		return binIdx
	}

	if ctxIdxOffset == 27 {
		if binIdx == 0 {
			return p.CtxIdxIncMbType(ctxIdxOffset)
		}
		if binIdx == 1 {
			return 3
		}
		if binIdx == 2 {
			return p.CtxIdxIncPrior(ctxIdxOffset, binIdx)
		}
		return 5
	}

	if ctxIdxOffset == 36 {
		if binIdx < 2 {
			return binIdx
		}
		if binIdx == 2 {
			return p.CtxIdxIncPrior(ctxIdxOffset, binIdx)
		}
		return 3
	}

	if ctxIdxOffset == 40 || ctxIdxOffset == 47 {
		if binIdx == 0 {
			// TODO: Clause 9.3.3.1.1.7
			panic("Not yet implemented!")

		}
		if binIdx >= 4 {
			return 6
		}
		return binIdx + 2
	}

	if ctxIdxOffset == 54 {
		if binIdx == 0 {
			// TODO: Clause 9.3.3.1.1.6
			panic("Not yet implemented!")

		}
		if binIdx == 1 {
			return 4
		}
		return 5
	}

	if ctxIdxOffset == 60 {
		if binIdx == 0 {
			return p.CtxIdxIncMbQpDelta()
		}
		if binIdx == 1 {
			return 2
		}
		return 3
	}

	if ctxIdxOffset == 64 {
		if binIdx == 0 {
			return p.CtxIdxIncIntraChromaPredMode()
		}

		return 3
	}

	// 68, 69, 276 handled by return 0

	if ctxIdxOffset == 70 {
		// TODO: Clause 9.3.3.1.1.2
		panic("Not yet implemented!")
	}

	if ctxIdxOffset == 73 || ctxIdxOffset == 77 {
		if ctxIdxOffset == 73 {
			return p.CtxIdxIncCodedBlockPattern73(binIdx)
		} else {
			return p.CtxIdxIncCodedBlockPattern77(binIdx)
		}

	}

	if ctxIdxOffset == 399 {
		return p.CtxIdxIncTransformSizeFlag()
	}

	return 0
}

func (p *CabacParser) DecodeBin(ctxIdx uint, bypassFlag uint) uint {
	if bypassFlag == 1 {
		return p.DecodeBypass()
	}

	if ctxIdx == TERMINATE_CTX {
		return p.DecodeTerminate()
	}
	if ctxIdx > WELS_CONTEXT_COUNT {
		panic("this should not happen")
	}
	return p.DecodeDecision(&p.contexts[ctxIdx])
}

func (p *CabacParser) RenormD() {
	for p.codIRange < 256 {
		p.codIRange <<= 1
		p.codIOffset <<= 1
		p.codIOffset |= uint16(p.ReadBits(1))
	}
	if p.codIOffset >= p.codIRange {
		panic("this should not happen")
	}
}

func (p *CabacParser) DecodeDecision(elem *CabacElement) (binVal uint) {
	qCodIRangeIdx := (p.codIRange >> 6) & 3
	codIRangeLPS := g_kuiCabacRangeLps[elem.uiState][qCodIRangeIdx]
	p.codIRange = p.codIRange - uint16(codIRangeLPS)
	if p.codIOffset >= p.codIRange {
		binVal = 1 - elem.uiMPS
		p.codIOffset -= p.codIRange
		p.codIRange = uint16(codIRangeLPS)

		if elem.uiState == 0 {
			elem.uiMPS = 1 - elem.uiMPS
		}

		elem.uiState = uint(g_kuiStateTransTable[elem.uiState][0])

	} else {
		binVal = elem.uiMPS
		elem.uiState = uint(g_kuiStateTransTable[elem.uiState][1])
	}

	p.RenormD()

	return
}

func (p *CabacParser) DecodeBypass() (binVal uint) {
	p.codIOffset <<= 1
	p.codIOffset |= uint16(p.ReadBits(1))
	if p.codIOffset >= p.codIRange {
		binVal = 1
		p.codIOffset -= p.codIRange
	} else {
		binVal = 0
	}

	return
}

func (p *CabacParser) DecodeTerminate() (binVal uint) {
	p.codIRange -= 2

	if p.codIOffset >= p.codIRange {
		binVal = 1
	} else {
		binVal = 0
		p.RenormD()
	}

	return
}
