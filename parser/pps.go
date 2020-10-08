package parser

import "go.uber.org/zap"

func NewPPSParser(p *H264Parser) *PPSParser {
	return &PPSParser{
		p.GolombBitReader,
		p.log.Named("PPSParser"),
		p,
	}
}

type PPSParser struct {
	*GolombBitReader
	log  *zap.SugaredLogger
	h264 *H264Parser
}

type PPSInfo struct {
	Id    uint
	SPSId uint
	SPS   *SPSInfo

	EntropyCodingModeFlag uint
	PicOrderPresentFlag   uint
	NumSliceGroupsMinus1  uint

	SliceGroupMapType uint

	NumRefIdxL0DefaultActiveMinus1     uint
	NumRefIdxL1DefaultActiveMinus1     uint
	WeightedPredFlag                   uint
	WeightedBipredIdc                  uint
	PicInitQPMinus26                   int
	PicInitQSMinus26                   int
	ChromaQPIndexOffset                int
	DeblockingFilterControlPresentFlag uint
	ConstrainedIntraPredFlag           uint
	RedundantPicCntPresentFlag         uint

	Transform8x8ModeFlag uint
}

func (p *PPSParser) ParseInfo() *PPSInfo {
	s := &PPSInfo{}

	// pic_parameter_set_id
	s.Id = p.ReadUE()
	// seq_parameter_set_id
	s.SPSId = p.ReadUE()
	s.SPS = p.h264.SPSInfos[s.SPSId]
	s.EntropyCodingModeFlag = p.ReadBits(1)
	s.PicOrderPresentFlag = p.ReadBits(1)
	s.NumSliceGroupsMinus1 = p.ReadUE()

	// if (num_slice_groups_minus1 > 0)
	// TODO
	if s.NumSliceGroupsMinus1 > 0 {
		panic("Not implemented yet!")
	}

	s.NumRefIdxL0DefaultActiveMinus1 = p.ReadUE()
	s.NumRefIdxL1DefaultActiveMinus1 = p.ReadUE()
	s.WeightedPredFlag = p.ReadBits(1)
	s.WeightedBipredIdc = p.ReadBits(2)

	s.PicInitQPMinus26 = p.ReadSE()
	s.PicInitQSMinus26 = p.ReadSE()
	s.ChromaQPIndexOffset = p.ReadSE()

	s.DeblockingFilterControlPresentFlag = p.ReadBits(1)
	s.ConstrainedIntraPredFlag = p.ReadBits(1)
	s.RedundantPicCntPresentFlag = p.ReadBits(1)

	// TODO: Check for more data and then parse other stuff!

	if p.MoreRBSPData() {
		s.Transform8x8ModeFlag = p.ReadBits(1)
		p.ReadBits(1)
		p.ReadSE()
		//panic("Not yet implemented!")
		// TODO: More stuff here!!!
	}

	// trailing
	stopBit := p.ReadBits(1)
	if stopBit != 1 {
		p.log.Warnf("Did not encounter trailing stop bit at position: %d", p.Position())
	}

	return s
}
