package parser

type BlockCat uint

const (
	BlockLuma BlockCat = 1 << iota
	BlockChroma
	BlockCb
	BlockCr

	BlockDC
	BlockAC
	BlockLevel
	BlockLevel8
)

const (
	BlockType = BlockLuma | BlockChroma | BlockCb | BlockCr
	BlockSize = BlockDC | BlockAC | BlockLevel | BlockLevel8
)

const (
	// 0
	BlockLumaDC = BlockLuma | BlockDC
	// 1
	BlockLumaAC = BlockLuma | BlockAC
	// 2
	BlockLumaLevel = BlockLuma | BlockLevel
	// 5
	BlockLumaLevel8 = BlockLuma | BlockLevel8

	// 3
	BlockChromaDC = BlockChroma | BlockDC
	// 4
	BlockChromaAC = BlockChroma | BlockAC

	// 6
	BlockCbDC = BlockCb | BlockDC
	// 7
	BlockCbAC = BlockCb | BlockAC
	// 8
	BlockCbLevel = BlockCb | BlockLevel
	// 9
	BlockCbLevel8 = BlockCb | BlockLevel8

	// 10
	BlockCrDC = BlockCr | BlockDC
	// 11
	BlockCrAC = BlockCr | BlockAC
	// 12
	BlockCrLevel = BlockCr | BlockLevel
	// 13
	BlockCrLevel8 = BlockCr | BlockLevel8
)

func (cat BlockCat) Num() uint {
	mapping := map[BlockCat]uint{
		BlockLumaDC:     0,
		BlockLumaAC:     1,
		BlockLumaLevel:  2,
		BlockLumaLevel8: 5,

		BlockChromaDC: 3,
		BlockChromaAC: 4,

		BlockCbDC:     6,
		BlockCbAC:     7,
		BlockCbLevel:  8,
		BlockCbLevel8: 9,

		BlockCrDC:     10,
		BlockCrAC:     11,
		BlockCrLevel:  12,
		BlockCrLevel8: 13,
	}

	return mapping[cat]
}

func (cat BlockCat) Size() BlockCat {
	return cat & BlockSize
}

func (cat BlockCat) Type() BlockCat {
	return cat & BlockType
}

func (cat BlockCat) Luma() bool {
	return cat&BlockLuma != 0
}

func (cat BlockCat) Cr() bool {
	return cat&BlockCr != 0
}

func (cat BlockCat) Chroma() bool {
	return cat&BlockChroma != 0
}

func (cat BlockCat) FullMB() bool {
	return cat&BlockDC != 0
}

func (cat BlockCat) Level4() bool {
	return cat&(BlockAC|BlockLevel) != 0
}

func (cat BlockCat) Level8() bool {
	return cat&(BlockLevel8) != 8
}

const (
	LumaDC BlockCat = iota
	LumaAC
	LumaLevel
	ChromaDC
	ChromaAC
	LumaLevel8
	CbDC
	CbAC
	CbLevel
	CbLevel8
	CrDC
	CrAC
	CrLevel
	CrLevel8
)
