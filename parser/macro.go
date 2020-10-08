package parser

import "fmt"

type MBPartPredMode uint

const (
	PartPred_Unknown MBPartPredMode = iota
	Intra_4x4
	Intra_8x8
	Intra_16x16
	Pred_L0
	Pred_L1
	BiPred
	Direct
)

type Intra16x16PredMode uint

const (
	Vertical Intra16x16PredMode = iota
	Horizontal
	DC
	Plane
	IntraPred_Unknown
)

type CodedBlockPatternChroma uint

const (
	AllCoeffs0 CodedBlockPatternChroma = iota
	AllACCoeffs0
	AllDCCoeffs0
	PatternChroma_Unknown
)

type GeneralIType uint

const (
	G_I_NxN GeneralIType = iota
	G_I_PCM
	G_Intra_16x16
)

type AllMBType interface {
	// FromInt(val uint) IMBType should exist
	MbPartPredMode(mbPartIdx uint) MBPartPredMode
}

type IMBType interface {
	AllMBType
	IntraPredMode() Intra16x16PredMode
	CodedBlockPatternChroma() CodedBlockPatternChroma
	CodedBlockPatternLuma() uint
	GeneralIType() GeneralIType
}

type PMBType interface {
	AllMBType
	NumMbPart() uint
	MbPartWidth() uint
	MbPartHeight() uint
}

type MBAddr int

const MBUnavailable MBAddr = -1

type LumaBlkIdx int

const LumaBlkUnavailable LumaBlkIdx = -1

func NewMacroBlock() *MacroBlock {
	mb := &MacroBlock{}
	mb.lumaResidual.cat = BlockLuma
	mb.cbResidual.cat = BlockCb
	mb.crResidual.cat = BlockCr
	mb.chromaResidual[0].cat = BlockChroma
	mb.chromaResidual[1].cat = BlockChroma
	return mb
}

type MacroBlock struct {
	Addr     MBAddr
	sliceHdr *SliceHeader
	qpVal    int

	mbType               uint
	qpDelta              int
	transformSize8x8Flag uint
	pcmSampleLuma        []uint
	pcmSampleChroma      []uint
	codedBlockPattern    uint

	// Prediction Semantics
	// Clause 7.4.5.1
	prevIntraPredModeFlag []uint
	remIntraPredMode      []uint
	intraChromaPredMode   uint

	// Residual block Semantics
	lumaResidual   ResidualData
	chromaResidual [2]ResidualData
	cbResidual     ResidualData
	crResidual     ResidualData
	// Clause 7.4.5.3.3
	codedBlockFlag           uint
	significantCoeffFlag     []uint
	lastSignificantCoeffFlag []uint
	coeffAbsLevelMinus1      []uint
	coeffSignFlag            []uint
}

func (mb *MacroBlock) IMBType() IMBType {
	return NewIMBType(mb)
}

func (mb *MacroBlock) CodedBlockPatternLuma() uint {
	return mb.codedBlockPattern % 16
}

func (mb *MacroBlock) CodedBlockPatternChroma() CodedBlockPatternChroma {
	return CodedBlockPatternChroma(mb.codedBlockPattern / 16)
}

func (mb *MacroBlock) Position() (uint, uint) {
	sps := mb.sliceHdr.PPS.SPS
	return uint(mb.Addr) % sps.PicWidthInMbs(), uint(mb.Addr) / sps.PicWidthInMbs()
}

func (mb *MacroBlock) GetResidual(cat BlockCat) *ResidualData {
	if cat.Luma() {
		return &mb.lumaResidual
	}
	if cat.Cr() {
		return &mb.crResidual
	}
	return &mb.cbResidual
}

func InverseRasterScan(a, b, c, d, e uint) uint {
	if e == 0 {
		return (a % (d / b)) * b
	}

	return (a / (d / b)) * c
}

func (mb *MacroBlock) LumaPosition() (uint, uint) {
	sps := mb.sliceHdr.PPS.SPS
	return InverseRasterScan(uint(mb.Addr), 16, 16, sps.PicWidthInSamplesL(), 0), InverseRasterScan(uint(mb.Addr), 16, 16, sps.PicWidthInSamplesL(), 1)
}

func (mb *MacroBlock) String() string {
	x, y := mb.Position()
	return fmt.Sprintf("<MB %s @ %d %d (%d)>", mb.IMBType(), x, y, mb.qpVal)
}
