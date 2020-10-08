package parser

import "fmt"

type IMBTypeConst uint

const (
	I_NxN IMBTypeConst = iota
	I_16x16_0_0_0
	I_16x16_1_0_0
	I_16x16_2_0_0
	I_16x16_3_0_0
	I_16x16_0_1_0
	I_16x16_1_1_0
	I_16x16_2_1_0
	I_16x16_3_1_0
	I_16x16_0_2_0
	I_16x16_1_2_0
	I_16x16_2_2_0
	I_16x16_3_2_0
	I_16x16_0_0_1
	I_16x16_1_0_1
	I_16x16_2_0_1
	I_16x16_3_0_1
	I_16x16_0_1_1
	I_16x16_1_1_1
	I_16x16_2_1_1
	I_16x16_3_1_1
	I_16x16_0_2_1
	I_16x16_1_2_1
	I_16x16_2_2_1
	I_16x16_3_2_1
	I_PCM
)

func NewIMBType(mb *MacroBlock) IMBType {
	return &IMBTypeImpl{mb}
}

type IMBTypeImpl struct {
	mb *MacroBlock
}

func (t IMBTypeConst) MBType() uint {
	return uint(t)
}

const (
	IIntraPredModeMask = 3
)

func (i *IMBTypeImpl) typeConst() IMBTypeConst {
	return IMBTypeConst(i.mb.mbType)
}

func (i *IMBTypeImpl) MbPartPredMode(mbPartIdx uint) MBPartPredMode {
	t := i.typeConst()
	switch t {
	case I_NxN:
		if i.mb.transformSize8x8Flag == 1 {
			return Intra_8x8
		}
		return Intra_4x4
	case I_PCM:
		return PartPred_Unknown
	default:
		return Intra_16x16
	}
}

func (i *IMBTypeImpl) IntraPredMode() Intra16x16PredMode {
	t := i.typeConst()
	switch t {
	case I_NxN:
	case I_PCM:
		return IntraPred_Unknown
	}

	val := t.MBType() - 1

	return Intra16x16PredMode(val & IIntraPredModeMask)
}

func (i *IMBTypeImpl) CodedBlockPatternChroma() CodedBlockPatternChroma {
	t := i.typeConst()
	switch t {
	case I_NxN:
		return i.mb.CodedBlockPatternChroma()
	case I_PCM:
		return PatternChroma_Unknown
	}

	val := t.MBType() - 1
	retVal := (val / 4) % 3
	return CodedBlockPatternChroma(retVal)
}

func (i *IMBTypeImpl) CodedBlockPatternLuma() uint {
	t := i.typeConst()
	switch t {
	case I_NxN:
		return i.mb.CodedBlockPatternLuma()
	case I_PCM:
		return 0xff
	}
	if t > 12 {
		return 15
	}

	return 0
}

func (i *IMBTypeImpl) GeneralIType() GeneralIType {
	t := i.typeConst()
	switch t {
	case I_NxN:
		return G_I_NxN
	case I_PCM:
		return G_I_PCM
	default:
		return G_Intra_16x16
	}
}

func (i *IMBTypeImpl) String() string {
	t := "I_PCM"
	switch i.MbPartPredMode(0) {
	case Intra_4x4:
		t = "Intra_4x4"
	case Intra_8x8:
		t = "Intra_8x8"
	case Intra_16x16:
		t = "Intra_16x16"
	}
	return fmt.Sprintf("%s (%d)", t, i.typeConst().MBType())
}
