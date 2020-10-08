package parser

type BaseCabacSE struct {
	p            *CabacParser
	bin          Binarization
	maxBinIdx    uint
	ctxIdxOffset uint
}

func (b *BaseCabacSE) Binarization() Binarization {
	return b.bin
}

func (b *BaseCabacSE) MaxBinIdx() uint {
	return b.maxBinIdx
}

func (b *BaseCabacSE) GetCtxIdx(binIdx uint) uint {
	return b.ctxIdxOffset
}

func (b *BaseCabacSE) Bypass() bool {
	return false
}

func (b *BaseCabacSE) SetParser(p *CabacParser) {
	b.p = p
}

func NewRegSE(maxBinIdx, ctxIdxOffset uint, bin Binarization) *RegularCabacSE {
	return &RegularCabacSE{
		BaseCabacSE{
			maxBinIdx:    maxBinIdx,
			ctxIdxOffset: ctxIdxOffset,
			bin:          bin,
		},
	}
}

type RegularCabacSE struct {
	BaseCabacSE
}

/* Clause 9.3.3.1 */
func (b *RegularCabacSE) GetCtxIdx(binIdx uint) uint {
	return b.p.GetCtxIdxInc(binIdx, b.ctxIdxOffset) + b.ctxIdxOffset
}

func NewFlagSE(ctxIdxOffset uint) *RegularCabacSE {
	return NewRegSE(0, ctxIdxOffset, FlagBin)
}

var MBTypeSE = NewRegSE(6, 3, MbTypeIBin)
var TransformSizeFlagSE = NewFlagSE(399)
var CodedBlockPatternLumaSE = NewRegSE(3, 73, NewFLBin(15))
var CodedBlockPatternChromaSE = NewRegSE(1, 77, NewTUBin(2))
var MBQPDeltaSE = NewRegSE(2, 60, NewUBin())
var IntraChromaPredModeSE = NewRegSE(1, 64, NewTUBin(3))
var PrevIntraPredModeFlagSE = NewFlagSE(68)
var RemIntraPredModeSE = NewRegSE(0, 69, NewFLBin(7))
var MBFieldDecodingFlagSE = NewFlagSE(70)
var EndOfSliceSE = NewFlagSE(TERMINATE_CTX)

func NewBlockSE(maxBinIdx uint, blockCatOffset [14]uint, ctxIdxOffset, ctxIdxInc func(*BlockCabacSE) uint, bin Binarization) *BlockCabacSE {
	return &BlockCabacSE{
		BaseCabacSE{
			maxBinIdx:    maxBinIdx,
			ctxIdxOffset: 0,
			bin:          bin,
		},
		nil,
		0,
		blockCatOffset,
		ctxIdxInc,
		ctxIdxOffset,
	}
}

type BlockCabacSE struct {
	BaseCabacSE
	block          *ResidualBlock
	levelListIdx   uint
	blockCatOffset [14]uint
	ctxIdxInc      func(*BlockCabacSE) uint
	ctxIdxOffset   func(*BlockCabacSE) uint
}

func (b *BlockCabacSE) GetCtxIdx(binIdx uint) uint {
	return b.ctxIdxOffset(b) + b.ctxIdxInc(b) + b.blockCatOffset[b.block.cat.Num()]
}

func NewBFlagSE(blockCatOffset [14]uint, ctxIdxOffset, ctxIdxInc func(*BlockCabacSE) uint) *BlockCabacSE {
	return NewBlockSE(0, blockCatOffset, ctxIdxOffset, ctxIdxInc, FlagBin)
}

func CBFOffset(b *BlockCabacSE) uint {
	cat := b.block.cat.Num()
	if cat == 5 || cat == 9 || cat == 13 {
		return 1012
	}
	if cat < 5 {
		return 85
	}
	if cat < 9 {
		return 460
	}
	if cat < 13 {
		return 472
	}
	return 1024
}

func CBFGetTransBlock(b *BlockCabacSE, nt NType) (mbAddrN MBAddr, transBlockN *ResidualBlock) {
	cat := b.block.cat
	blkIdx := b.block.blkIdx
	mbAddrN, blkIdxN := b.p.s.NeighBlocksN(nt, blkIdx, cat)
	transBlockN = nil
	if mbAddrN == MBUnavailable || blkIdxN == LumaBlkUnavailable {
		return
	}
	mb := b.p.s.MacroBlocks[mbAddrN]
	residual := mb.GetResidual(cat)
	if cat.FullMB() {
		if cat.Chroma() {
			panic("Not implemented yet")
		}
		if mb.IMBType().MbPartPredMode(0) != Intra_16x16 {
			return
		}
		transBlockN = residual.GetBlock(cat, 0)
		return
	}
	if cat.Level4() {
		if ((mb.CodedBlockPatternLuma() >> (blkIdxN >> 2)) & 1) != 0 {
			if mb.transformSize8x8Flag == 0 {
				transBlockN = residual.GetBlock(cat.Type()|BlockLevel, uint(blkIdxN))
			} else {
				transBlockN = residual.GetBlock(cat.Type()|BlockLevel8, uint(blkIdxN)>>2)
			}
		}
		return
	}
	if cat.Level8() {
		if ((mb.CodedBlockPatternLuma()>>blkIdxN)&1) != 0 && mb.transformSize8x8Flag == 1 {
			transBlockN = residual.GetBlock(cat, uint(blkIdxN))
		}
	}

	return
}

func CBFCondFlag(b *BlockCabacSE, nt NType) uint {
	mbAddrN, transBlockN := CBFGetTransBlock(b, nt)
	if mbAddrN == MBUnavailable /* Check for Intra! */ {
		return 1
	}
	if transBlockN == nil {
		return 0
	}
	panic("Not yet implemented")
}

func CBFInc(b *BlockCabacSE) uint {
	condFlagA := CBFCondFlag(b, NA)
	condFlagB := CBFCondFlag(b, NB)

	return condFlagA + 2*condFlagB
}

func SCFOffset(b *BlockCabacSE) uint {
	cat := b.block.cat.Num()
	switch cat {
	case 5:
		return 402
	case 9:
		return 660
	case 13:
		return 718
	}
	if cat < 5 {
		return 105
	}
	if cat < 9 {
		return 484
	}
	return 528
}

func LSCFOffset(b *BlockCabacSE) uint {
	cat := b.block.cat.Num()
	switch cat {
	case 5:
		return 417
	case 9:
		return 690
	case 13:
		return 748
	}
	if cat < 5 {
		return 166
	}
	if cat < 9 {
		return 572
	}
	return 616
}

func SCFInc(b *BlockCabacSE) uint {
	cat := b.block.cat
	if cat.Size() != BlockLevel8 && cat != BlockChromaDC {
		return b.levelListIdx
	}
	if cat == BlockChromaDC {
		val := b.levelListIdx / b.p.s.h264.SPS.NumC8x8()
		if val > 2 {
			return 2
		}
	}

	return kuiSignificantCoeffFlagOffset8x8[0][b.levelListIdx]
}

func LSCFInc(b *BlockCabacSE) uint {
	cat := b.block.cat
	if cat.Size() != BlockLevel8 && cat != BlockChromaDC {
		return b.levelListIdx
	}
	if cat == BlockChromaDC {
		val := b.levelListIdx / b.p.s.h264.SPS.NumC8x8()
		if val > 2 {
			return 2
		}
	}

	return kuiLastSignificantCoeffFlagOffset8x8[b.levelListIdx]
}

var CodedBlockFlagSE = NewBFlagSE(kuiCtxIdxBlockCatOffset[0], CBFOffset, CBFInc)
var SignificantCoeffFlagSE = NewBFlagSE(kuiCtxIdxBlockCatOffset[1], SCFOffset, SCFInc)
var LastSignificantCoeffFlagSE = NewBFlagSE(kuiCtxIdxBlockCatOffset[2], LSCFOffset, LSCFInc)
