package parser

import (
	"go.uber.org/zap"
)

func NewSliceParser(p *H264Parser) *SliceParser {
	return &SliceParser{
		GolombBitReader: p.GolombBitReader,
		log:             p.log.Named("PPSParser"),
		h264:            p,
	}
}

type SliceParser struct {
	*GolombBitReader
	log  *zap.SugaredLogger
	h264 *H264Parser

	h *SliceHeader

	CurrMbAddr  uint
	CurrMb      *MacroBlock
	PrevMbAddr  MBAddr
	MacroBlocks map[MBAddr]*MacroBlock
	unitToGroup map[uint]uint
	MbToGroup   map[uint]uint
	CurrQP      int
}

type SliceHeader struct {
	FirstMbInSlice uint
	SliceType      uint
	PicParamSetId  uint

	PPS *PPSInfo

	FrameNum       uint
	IdrPicId       uint
	PicOrderCntLsb uint

	DecRefPicMarking *DecRefPicMarking

	SliceQPDelta int

	DisableDeblockingFilterIdc uint
	SliceAlphaC0OffsetDiv2     int
	SliceBetaOffsetDiv2        int
}

func (p *SliceParser) ParseHeader() *SliceHeader {
	h := &SliceHeader{}
	p.h = h

	h.FirstMbInSlice = p.ReadUE()
	h.SliceType = p.ReadUE()
	h.PicParamSetId = p.ReadUE()

	h.PPS = p.h264.PPSInfos[h.PicParamSetId]

	maxFrameBits := p.h264.SPS.Log2MaxFrameNumMinus4 + 4
	h.FrameNum = p.ReadBits(uint8(maxFrameBits))

	h.IdrPicId = p.ReadUE()
	if p.h264.SPS.PicOrderCntType == 0 {
		maxCntBits := p.h264.SPS.Log2MaxPicOrderCntLsbMinus4 + 4
		h.PicOrderCntLsb = p.ReadBits(uint8(maxCntBits))
	} else {
		p.log.Error("Cannot handle pic order cnt of type non zero!")
	}

	// TODO More shit here, that's currently not used
	// if (nal_ref_idc != 0)
	h.DecRefPicMarking = p.ParseDecRefPicMarking()

	h.SliceQPDelta = int(p.ReadSE())

	if h.PPS.DeblockingFilterControlPresentFlag != 0 {
		h.DisableDeblockingFilterIdc = p.ReadUE()
		if h.DisableDeblockingFilterIdc != 1 {
			h.SliceAlphaC0OffsetDiv2 = p.ReadSE()
			h.SliceBetaOffsetDiv2 = p.ReadSE()
		}
	}

	return h
}

type DecRefPicMarking struct {
	NoOutputOfPriorPicsFlag uint
	LongTermReferenceFlag   uint
}

func (p *SliceParser) ParseDecRefPicMarking() *DecRefPicMarking {
	r := &DecRefPicMarking{}

	r.NoOutputOfPriorPicsFlag = p.ReadBits(1)
	r.LongTermReferenceFlag = p.ReadBits(1)

	return r
}

func (p *SliceParser) CalculateSliceMap() {
	p.unitToGroup = map[uint]uint{}
	if p.h.PPS.NumSliceGroupsMinus1 == 0 {
		for i := uint(0); i < p.h264.SPS.PicSizeInMapUnits(); i++ {
			p.unitToGroup[i] = 0
		}
	} else {
		panic("Not implemented yet!")
	}
	p.CalculateMbMap()
}

func (p *SliceParser) CalculateMbMap() {
	p.MbToGroup = map[uint]uint{}
	for i := uint(0); i < p.h264.SPS.PicSizeInMbs(); i++ {
		res := uint(0)
		if p.h264.SPS.FrameMbsOnlyFlag == 1 {
			res = p.unitToGroup[i]
		} else {
			// TODO MBAFF
			res = p.unitToGroup[(i/(2*p.h264.SPS.PicWidthInMbs()))*p.h264.SPS.PicWidthInMbs()+(i%p.h264.SPS.PicWidthInMbs())]
		}
		p.MbToGroup[i] = res
	}
}

func (p *SliceParser) NextMbAddress(addr MBAddr) MBAddr {
	i := addr + 1
	for uint(i) < p.h264.SPS.PicSizeInMbs() && p.MbToGroup[uint(i)] != p.MbToGroup[uint(addr)] {
		i++
	}
	return i
}

type SliceData struct {
}

func (p *SliceParser) ParseSliceData() *SliceData {
	d := &SliceData{}
	p.MacroBlocks = map[MBAddr]*MacroBlock{}
	p.CalculateSliceMap()
	p.CurrQP = p.h.PPS.PicInitQPMinus26 + 26 + p.h.SliceQPDelta

	if p.h.PPS.EntropyCodingModeFlag != 0 {
		p.Align()
	}

	p.PrevMbAddr = MBUnavailable
	mbaffFrameFlag := uint(0)

	p.CurrMbAddr = p.h.FirstMbInSlice * (1 + mbaffFrameFlag)
	moreDataFlag := 1
	prevMbSkipped := 0
	c := NewCabacParser(p.h264)
	c.Initialize()
	c.InitializeDecodeEngine()

	for {
		// if (slice_type != I && slice_type != SI)
		// TODO

		if moreDataFlag != 0 {
			if mbaffFrameFlag != 0 && (p.CurrMbAddr%2 == 0 || (p.CurrMbAddr%2 == 1 && prevMbSkipped != 0)) {
				// TODO: What is this used for?
				c.ParseMbFieldDecodingFlag()
			}
			p.ParseMacroblockLayer(c)
		}

		if p.h.PPS.EntropyCodingModeFlag == 0 {
			moreDataFlag = 0
			if p.MoreRBSPData() {
				moreDataFlag = 1
			}
		} else {
			// if (slice_type != I && slice_type != SI)
			// 	prevMbSkipped = mbSkipFlag
			if mbaffFrameFlag != 0 && p.CurrMbAddr%2 == 0 {
				moreDataFlag = 1
			} else {
				endOfSliceFlag := c.ParseEndOfSliceFlag()
				moreDataFlag = 1
				if endOfSliceFlag != 0 {
					moreDataFlag = 0
				}
			}
		}

		p.PrevMbAddr = MBAddr(p.CurrMbAddr)
		p.CurrMbAddr = uint(p.NextMbAddress(p.PrevMbAddr))

		if moreDataFlag == 0 {
			break
		}
	}

	return d
}

func (p *SliceParser) ParseMacroblockLayer(c *CabacParser) {
	p.CurrMb = NewMacroBlock()
	p.CurrMb.Addr = MBAddr(p.CurrMbAddr)
	p.CurrMb.sliceHdr = p.h
	p.MacroBlocks[MBAddr(p.CurrMbAddr)] = p.CurrMb
	p.CurrMb.mbType = c.ParseMBType()
	imbType := p.CurrMb.IMBType()

	// PCM
	if imbType.GeneralIType() == G_I_PCM {
		// Shorter version of:
		/*
			while (!byte_aligned())
				pcm_alignment_zero_bit
		*/
		c.Align()

		pcm_sample_luma := [256]uint{}
		for i := 0; i < 256; i++ {
			pcm_sample_luma[i] = p.ReadBits(uint8(p.h264.SPS.BitDepthY()))
		}
		p.CurrMb.pcmSampleLuma = pcm_sample_luma[:]
		cNum := 2 * p.h264.SPS.MbHeightC() * p.h264.SPS.MbWidthC()
		pcm_sample_chroma := make([]uint, cNum)
		for i := uint(0); i < cNum; i++ {
			pcm_sample_chroma[i] = p.ReadBits(uint8(p.h264.SPS.BitDepthC()))
		}
		p.CurrMb.pcmSampleChroma = pcm_sample_chroma

	} else {
		noSubMbPartSizeLessThan8x8Flag := 1
		//pmbType := p.CurrMb.PMBType()
		if imbType.GeneralIType() != G_I_NxN && imbType.MbPartPredMode(0) != Intra_16x16 /* TODO: More here! */ {
			// Probably anything not I related?
			// TODO: even more stuff here
			panic("Not implemented yet!")
		} else {
			// Intra_YxY mode

			if p.h.PPS.Transform8x8ModeFlag == 1 && imbType.GeneralIType() == G_I_NxN {
				p.CurrMb.transformSize8x8Flag = c.ParseTransformSize8x8Flag()
			}
			// mb_pred
			p.ParseMbPred(c)
		}

		if imbType.MbPartPredMode(0) != Intra_16x16 {
			p.CurrMb.codedBlockPattern = c.ParseCodedBlockPattern(p.h264.SPS.ChromaArrayType())
			if p.CurrMb.CodedBlockPatternLuma() > 0 && p.h.PPS.Transform8x8ModeFlag == 1 && imbType.GeneralIType() != G_I_NxN && noSubMbPartSizeLessThan8x8Flag == 1 /* TODO: More conditions here */ {
				p.CurrMb.transformSize8x8Flag = c.ParseTransformSize8x8Flag()
			}
		}
		if imbType.MbPartPredMode(0) == Intra_16x16 || p.CurrMb.CodedBlockPatternChroma() > 0 || p.CurrMb.CodedBlockPatternLuma() > 0 {
			p.CurrMb.qpDelta = c.ParseMbQpDelta()
			p.ParseResidual(0, 15, c)
		}
	}

	p.CurrQP += p.CurrMb.qpDelta
	p.CurrMb.qpVal = p.CurrQP

	p.log.Debugf("%s", p.CurrMb)
}

func (p *SliceParser) ParseMbPred(c *CabacParser) {
	imbType := p.CurrMb.IMBType()
	mbPartPred := imbType.MbPartPredMode(0)
	if mbPartPred == Intra_4x4 || mbPartPred == Intra_8x8 || mbPartPred == Intra_16x16 {
		if mbPartPred == Intra_4x4 || mbPartPred == Intra_8x8 {
			num := 16
			if mbPartPred == Intra_8x8 {
				num = 4
			}
			p.CurrMb.prevIntraPredModeFlag = make([]uint, num)
			p.CurrMb.remIntraPredMode = make([]uint, num)
			for blkIdx := 0; blkIdx < num; blkIdx++ {
				p.CurrMb.prevIntraPredModeFlag[blkIdx] = c.ParsePrevIntraPredModeFlag()
				if p.CurrMb.prevIntraPredModeFlag[blkIdx] == 0 {
					p.CurrMb.remIntraPredMode[blkIdx] = c.ParseRemIntraPredMode()
				}
			}
		}
		if p.h264.SPS.ChromaArrayType() == 1 || p.h264.SPS.ChromaArrayType() == 2 {
			p.CurrMb.intraChromaPredMode = c.ParseIntraChromaPredMode()
		}
	} else if mbPartPred != Direct {
		panic("Not implemented yet!")
	}
}

func (p *SliceParser) ParseResidual(start, end uint, c *CabacParser) {
	// panic("Not implemented yet!")
	// We assume we are always parsing CABAC. As such we can skip this first few lines here:
	/*
		if (!entropy_coding_mode_flag)
			residual_block = residual_block_cavlc
		else
			residual_block = residual_block_cabac
	*/
	p.CurrMb.lumaResidual = *p.ParseResidualLuma(start, end, BlockLuma, c)

	if p.h264.SPS.ChromaArrayType() == 1 || p.h264.SPS.ChromaArrayType() == 2 {
		for iCbCr := 0; iCbCr < 2; iCbCr++ {
			residual := &p.CurrMb.chromaResidual[iCbCr]
			if (p.CurrMb.CodedBlockPatternChroma()&3) != 0 && start == 0 {
				p.ParseResidualBlock(residual.GetBlock(BlockDC, 0), 0, 4*p.h264.SPS.NumC8x8()-1, 4*p.h264.SPS.NumC8x8(), c)
			}
		}

		for iCbCr := 0; iCbCr < 2; iCbCr++ {
			residual := &p.CurrMb.chromaResidual[iCbCr]
			for i8x8 := uint(0); i8x8 < p.h264.SPS.NumC8x8(); i8x8++ {
				for i4x4 := uint(0); i4x4 < 4; i4x4++ {
					if (p.CurrMb.CodedBlockPatternChroma() & 2) != 0 {
						p.ParseResidualBlock(residual.GetBlock(BlockAC, i8x8*4+i4x4), Max(0, start-1), end-1, 15, c)
					}
				}
			}
		}
	} else if p.h264.SPS.ChromaArrayType() == 3 {
		panic("Not implemented yet")
	}
}

func Max(a, b uint) uint {
	if a > b {
		return a
	}

	return b
}

func (p *SliceParser) ParseResidualLuma(startIdx, endIdx uint, cat BlockCat, c *CabacParser) *ResidualData {
	imbType := p.CurrMb.IMBType()
	mbPartPred := imbType.MbPartPredMode(0)

	ret := &ResidualData{cat: cat}

	if startIdx == 0 && mbPartPred == Intra_16x16 {
		p.ParseResidualBlock(ret.GetBlock(BlockDC, 0), 0, 15, 16, c)
	}
	for i8x8 := 0; i8x8 < 4; i8x8++ {
		if p.CurrMb.transformSize8x8Flag == 0 || p.h.PPS.EntropyCodingModeFlag == 0 {
			for i4x4 := 0; i4x4 < 4; i4x4++ {
				blkIdx := uint(i8x8*4 + i4x4)
				if p.CurrMb.CodedBlockPatternLuma()&(1<<i8x8) != 0 {
					if mbPartPred == Intra_16x16 {
						p.ParseResidualBlock(ret.GetBlock(BlockAC, blkIdx), Max(0, startIdx-1), endIdx-1, 15, c)
					} else {
						p.ParseResidualBlock(ret.GetBlock(BlockLevel, blkIdx), startIdx, endIdx, 16, c)
					}
				} else {
					//panic("Not implemented yet")
				}
			}
		} else if p.CurrMb.CodedBlockPatternLuma()&(1<<i8x8) != 0 {
			p.ParseResidualBlock(ret.GetBlock(BlockLevel8, uint(i8x8)), 4*startIdx, 4*endIdx+3, 64, c)
		}
	}

	return ret
}

func (p *SliceParser) ParseResidualBlock(coeffLevel *ResidualBlock, startIdx, endIdx, maxNumCoeff uint, c *CabacParser) {
	coded_block_flag := uint(1)
	if maxNumCoeff != 64 || p.h264.SPS.ChromaArrayType() == 3 {
		coded_block_flag = c.ParseCodedBlockFlag(coeffLevel)
	}
	for i := uint(0); i < maxNumCoeff; i++ {
		coeffLevel.level[i] = 0
	}
	if coded_block_flag != 0 {
		significant_coeff_flag := make([]uint, 64)
		numCoeff := endIdx + 1
		i := startIdx
		for i < numCoeff-1 {
			significant_coeff_flag[i] = c.ParseSignificantCoeffFlag(coeffLevel)
			if significant_coeff_flag[i] != 0 {
				last_significant_coeff_flag := c.ParseLastSignificantCoeffFlag(coeffLevel)
				if last_significant_coeff_flag != 0 {
					numCoeff++
				}
			}
			i++
		}
	}
}
