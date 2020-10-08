package parser

type NType uint

const (
	NA NType = iota
	NB
	NC
	ND
)

/* Table 6-2 */
func InOutAssign(nt NType) (xD int, yD int) {
	switch nt {
	case NA:
		return -1, 0
	case NB:
		return 0, -1
	case NC:
		// TODO: predPartWidth
		return 0, -1
	case ND:
		return -1, -1
	}

	return 0, 0
}

/* Clause 6.4.11.1 */
func (p *SliceParser) NeighMacroBlock() (mbAddrA MBAddr, mbAddrB MBAddr) {
	addrA, addrB, _, _ := p.NeighBlocks(0, BlockLumaDC)
	return addrA, addrB
}

func (p *SliceParser) NeighBlocks(blkIdx uint, cat BlockCat) (mbAddrA MBAddr, mbAddrB MBAddr, lumaBlkIdxA LumaBlkIdx, lubaBlkIdxB LumaBlkIdx) {
	mbAddrA, lumaBlkIdxA = p.NeighBlocksN(NA, blkIdx, cat)
	mbAddrB, lubaBlkIdxB = p.NeighBlocksN(NB, blkIdx, cat)
	return
}

func (p *SliceParser) NeighBlocksN(nt NType, blkIdx uint, cat BlockCat) (MBAddr, LumaBlkIdx) {
	xD, yD := InOutAssign(nt)
	x, y := p.NeighBlocksXY(blkIdx, cat)
	xN := x + xD
	yN := y + yD
	addr, xW, yW := p.NeighLocation(xN, yN, cat)
	idx := LumaBlkUnavailable
	if addr != MBUnavailable {
		idx = LumaBlkIdxCat(xW, yW, cat)
	}
	return addr, idx
}

func (p *SliceParser) NeighBlocksXY(blkIdx uint, cat BlockCat) (x, y int) {
	idx := int(blkIdx)
	if cat.FullMB() {
		return 0, 0
	}
	if cat.Level8() {
		return (idx % 2) * 8, (idx / 2) * 8
	}
	if cat.Level4() {
		if cat.Chroma() {
			return InverseLevel4ChromaScan(blkIdx)
		}
		return InverseLevel4LumaScan(blkIdx)
	}

	return 0, 0
}

func InverseLevel4ScanHelper(blkIdx uint, e uint) uint {
	return InverseRasterScan(blkIdx/4, 8, 8, 16, e) + InverseRasterScan(blkIdx%4, 4, 4, 8, e)
}

func InverseLevel4LumaScan(blkIdx uint) (int, int) {
	return int(InverseLevel4ScanHelper(blkIdx, 0)), int(InverseLevel4ScanHelper(blkIdx, 1))
}

func InverseLevel8LumaScan(blkIdx uint) (int, int) {
	return int(InverseRasterScan(blkIdx, 8, 8, 16, 0)), int(InverseRasterScan(blkIdx, 8, 8, 16, 1))
}

func InverseLevel4ChromaScan(blkIdx uint) (int, int) {
	return int(InverseRasterScan(blkIdx, 4, 4, 8, 0)), int(InverseRasterScan(blkIdx, 4, 4, 8, 1))
}

/* Clause 6.4.12 */

func (p *SliceParser) NeighLocation(xN int, yN int, cat BlockCat) (mbAddrN MBAddr, xW int, yW int) {
	// Luma
	maxW := 16
	maxH := maxW

	if !cat.Luma() {
		maxW = int(p.h264.SPS.MbWidthC())
		maxH = int(p.h264.SPS.MbHeightC())
	}

	// Non MBAFF for now
	return p.NLocationNonMBAFF(xN, yN, maxW, maxH)
}

/* Clause 6.4.12.1 */
func (p *SliceParser) NLocationNonMBAFF(xN int, yN int, maxW int, maxH int) (mbAddrN MBAddr, xW int, yW int) {
	xW = (xN + maxW) % maxW
	yW = (yN + maxH) % maxH
	mbAddrN = MBUnavailable

	a, b, c, d := p.NMBAddrAvail()
	if xN < 0 {
		if yN < 0 {
			mbAddrN = d
		} else if yN < maxH {
			mbAddrN = a
		}
	} else if xN < maxW {
		if yN < 0 {
			mbAddrN = b
		} else if yN < maxH {
			mbAddrN = MBAddr(p.CurrMbAddr)
		}
	} else {
		if yN < 0 {
			mbAddrN = c
		}
	}

	return
}

func LumaBlkIdxCat(xP, yP int, cat BlockCat) LumaBlkIdx {
	if !cat.Luma() {
		return Chroma4BlkIdx(xP, yP)
	}

	if cat.Level4() {
		return Luma4BlkIdx(xP, yP)
	}

	return Luma8BlkIdx(xP, yP)
}

/* Clause 6.4.13.1 */
func Luma4BlkIdx(xP, yP int) LumaBlkIdx {
	return LumaBlkIdx(8*(yP/8) + 4*(xP/8) + 2*((yP%8)/4) + ((xP % 8) / 4))
}

/* Clause 6.4.13.2 */
func Chroma4BlkIdx(xP, yP int) LumaBlkIdx {
	return LumaBlkIdx(2*(yP/4) + (xP / 4))
}

/* Clause 6.4.13.3 */
func Luma8BlkIdx(xP, yP int) LumaBlkIdx {
	return LumaBlkIdx(2*(yP/8) + (xP / 8))
}

/* Clause 6.4.9 */
func (p *SliceParser) NMBAddrAvail() (mbAddrA MBAddr, mbAddrB MBAddr, mbAddrC MBAddr, mbAddrD MBAddr) {
	currMbAddr := MBAddr(p.CurrMbAddr)
	PicWidthInMbs := MBAddr(p.h264.SPS.PicWidthInMbs())
	mbAddrA = currMbAddr - 1
	if !p.MBAvailable(mbAddrA) || currMbAddr%PicWidthInMbs == 0 {
		mbAddrA = MBUnavailable
	}
	mbAddrB = currMbAddr - PicWidthInMbs
	if !p.MBAvailable(mbAddrB) {
		mbAddrB = MBUnavailable
	}
	mbAddrC = currMbAddr - PicWidthInMbs + 1
	if !p.MBAvailable(mbAddrC) || (currMbAddr+1)%PicWidthInMbs == 0 {
		mbAddrC = MBUnavailable
	}
	mbAddrD = currMbAddr - PicWidthInMbs - 1
	if !p.MBAvailable(mbAddrD) || currMbAddr%PicWidthInMbs == 0 {
		mbAddrD = MBUnavailable
	}
	return
}

/* Clause 6.4.8 */
func (p *SliceParser) MBAvailable(mbAddr MBAddr) bool {
	return !(mbAddr < 0 || mbAddr > MBAddr(p.CurrMbAddr) || mbAddr < MBAddr(p.h.FirstMbInSlice))
}
