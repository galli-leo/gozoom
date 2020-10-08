package parser

type ctxIdxIncF func(uint, uint) uint

/* Clause 9.3.3.1.2 */
func (p *CabacParser) CtxIdxIncPrior(ctxIdxOffset uint, binIdx uint) uint {
	priorIdx := uint(1)
	if binIdx != 2 {
		priorIdx = 3
	}
	b := p.getBin(priorIdx)
	ret := kuiCtxIdxIncPrior[ctxIdxOffset][b]
	if ctxIdxOffset == 3 {
		ret++
	}
	return ret
}

/* Helper for Clauses 9.3.3.1.1.1-9.3.3.1.1.10 */
func (p *CabacParser) CtxIdxIncCondFlag(mbAddr MBAddr, cond func(*MacroBlock) bool) uint {
	if mbAddr == MBUnavailable {
		return 0
	}
	mb := p.s.MacroBlocks[mbAddr]
	if cond(mb) {
		return 1
	}
	return 0
}

func (p *CabacParser) CtxIdxIncCondFlagBlk(mbAddr MBAddr, blkIdx LumaBlkIdx, cond func(*MacroBlock, LumaBlkIdx) bool) uint {
	if mbAddr == MBUnavailable {
		return 0
	}
	mb := p.s.MacroBlocks[mbAddr]
	if cond(mb, blkIdx) {
		return 1
	}
	return 0
}

/* Clause 9.3.3.1.1.1 */
func (p *CabacParser) CtxIdxIncMbSkipFlag() uint {
	mbAddrA, mbAddrB := p.s.NeighMacroBlock()
	condTermFlagA := uint(1)
	condTermFlagB := condTermFlagA
	if mbAddrA == MBUnavailable {
		condTermFlagA = 0
	}
	if mbAddrB == MBUnavailable {
		condTermFlagB = 0
	}
	return condTermFlagA + condTermFlagB
}

/* Clause 9.3.3.1.1.3 */
func (p *CabacParser) CtxIdxIncMbType(ctxIdxOffset uint) uint {
	mbAddrA, mbAddrB := p.s.NeighMacroBlock()
	cond := func(mb *MacroBlock) bool {
		switch ctxIdxOffset {
		case 0:
			return false // TODO: MB_TYPE SI
		case 3:
			return mb.IMBType().GeneralIType() == G_I_NxN
		case 27:
			return false // TODO: MB_TYPE B_Skip / B_Direct_16x16
		}
		return false
	}
	condTermFlagA := p.CtxIdxIncCondFlag(mbAddrA, cond)
	condTermFlagB := p.CtxIdxIncCondFlag(mbAddrB, cond)
	return condTermFlagA + condTermFlagB
}

/* Clause 9.3.3.1.1.4 */
func (p *CabacParser) CtxIdxIncCodedBlockPattern73(binIdx uint) uint {
	mbAddrA, mbAddrB, lumaBlkIdxA, lumaBlkIdxB := p.s.NeighBlocks(binIdx, BlockLumaLevel8)
	cond := func(mb *MacroBlock, blkIdx LumaBlkIdx) bool {
		if mb.IMBType().GeneralIType() == G_I_PCM {
			return false
		}

		if blkIdx == LumaBlkUnavailable {
			return true
		}

		// TODO: MB_TYPE B_Skip / P_Skip
		if (mb.CodedBlockPatternLuma()>>uint(blkIdx))&1 != 0 {
			return false
		}

		if mb == p.s.CurrMb && p.getBin(uint(blkIdx)) != 0 {
			return false
		}

		return true
	}
	condTermFlagA := p.CtxIdxIncCondFlagBlk(mbAddrA, lumaBlkIdxA, cond)
	condTermFlagB := p.CtxIdxIncCondFlagBlk(mbAddrB, lumaBlkIdxB, cond)
	return condTermFlagA + 2*condTermFlagB
}

func (p *CabacParser) CtxIdxIncCodedBlockPattern77(binIdx uint) uint {
	mbAddrA, mbAddrB := p.s.NeighMacroBlock()
	cond := func(mb *MacroBlock) bool {
		if mb.IMBType().GeneralIType() == G_I_PCM {
			return true
		}

		// TODO: MB_TYPE B_Skip / P_Skip
		if mb.CodedBlockPatternChroma() == 0 && binIdx == 0 {
			return false
		}

		if mb.CodedBlockPatternChroma() != 2 && binIdx == 1 {
			return false
		}
		return true
	}
	condTermFlagA := p.CtxIdxIncCondFlag(mbAddrA, cond)
	condTermFlagB := p.CtxIdxIncCondFlag(mbAddrB, cond)
	add := uint(0)
	if binIdx == 1 {
		add = 4
	}
	return condTermFlagA + 2*condTermFlagB + add
}

/* Clause 9.3.3.1.1.5 */
func (p *CabacParser) CtxIdxIncMbQpDelta() uint {
	return p.CtxIdxIncCondFlag(p.s.PrevMbAddr, func(mb *MacroBlock) bool {
		// TODO: MB_TYPE P_Skip, B_Skip
		if mb.IMBType().GeneralIType() == G_I_PCM {
			return false
		}
		if mb.IMBType().MbPartPredMode(0) != Intra_16x16 && mb.CodedBlockPatternChroma() == 0 && mb.CodedBlockPatternLuma() == 0 {
			return false
		}
		if mb.qpDelta == 0 {
			return false
		}
		return true
	})
}

/* Clause 9.3.3.1.1.8 */
func (p *CabacParser) CtxIdxIncIntraChromaPredMode() uint {
	mbAddrA, mbAddrB := p.s.NeighMacroBlock()
	cond := func(mb *MacroBlock) bool {
		// TODO: Pred Mode
		if mb.IMBType().GeneralIType() == G_I_PCM || mb.intraChromaPredMode == 0 {
			return false
		}

		return true
	}
	condTermFlagA := p.CtxIdxIncCondFlag(mbAddrA, cond)
	condTermFlagB := p.CtxIdxIncCondFlag(mbAddrB, cond)
	return condTermFlagA + condTermFlagB
}

/* Clause 9.3.3.1.1.10 */
func (p *CabacParser) CtxIdxIncTransformSizeFlag() uint {
	mbAddrA, mbAddrB := p.s.NeighMacroBlock()
	cond := func(mb *MacroBlock) bool {
		return mb.transformSize8x8Flag != 0
	}
	condTermFlagA := p.CtxIdxIncCondFlag(mbAddrA, cond)
	condTermFlagB := p.CtxIdxIncCondFlag(mbAddrB, cond)
	return condTermFlagA + condTermFlagB
}
