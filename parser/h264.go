package parser

import (
	"io"

	"github.com/hashicorp/go-multierror"
	"go.uber.org/zap"
)

func NewH264Parser(r io.Reader, log *zap.SugaredLogger) *H264Parser {
	p := &H264Parser{r: r, log: log.Named("H264Parser")}
	p.Initialize()
	return p
}

type H264Parser struct {
	*GolombBitReader
	r      io.Reader
	br     BitReader
	annexb *AnnexBReader
	emuR   *EmuPreventionReader

	log *zap.SugaredLogger

	SPSInfos map[uint]*SPSInfo
	PPSInfos map[uint]*PPSInfo
	SPS      *SPSInfo

	sps   *SPSParser
	pps   *PPSParser
	slice *SliceParser
}

func (p *H264Parser) Initialize() {
	p.br = NewBitioReader(p.r)
	p.annexb = NewAnnexBReader(p.br, p.log)
	p.emuR = NewEmuPreventionReader(p.annexb, p.log)
	p.GolombBitReader = NewGolombReader(p.emuR)

	p.SPSInfos = map[uint]*SPSInfo{}
	p.PPSInfos = map[uint]*PPSInfo{}

	p.sps = NewSPSParser(p)
	p.pps = NewPPSParser(p)
	p.slice = NewSliceParser(p)
}

func (p *H264Parser) Parse() {
	p.log.Infof("Starting h264 Parsing")
	for p.annexb.Next() {
		info := p.annexb.Current()
		p.log.Debugf("Parsed NALU of type %s", info.Type)
		switch info.Type {
		case NALU_SPS:
			p.SPS = p.sps.ParseInfo()
			p.SPSInfos[p.SPS.Id] = p.SPS
			p.log.Debugf("Parsed SPS: %+v", p.SPS)
		case NALU_PPS:
			pps := p.pps.ParseInfo()
			p.PPSInfos[pps.Id] = pps
			p.log.Debugf("Parsed PPS: %+v", pps)
		case NALU_IDR:
			hdr := p.slice.ParseHeader()
			p.log.Debugf("Parsed Slice Header: %+v", hdr)
			data := p.slice.ParseSliceData()
			p.log.Debugf("Parsed Slice Data: %+v", data)
		}
		// ensure old buffer is gone
		p.GolombBitReader.Align()
		p.emuR.Reset()
	}
}

func (p *H264Parser) Error() error {
	return multierror.Append(multierror.Prefix(p.err, "H264Parser Errors"), multierror.Prefix(p.sps.err, "SPSParser Errors"), multierror.Prefix(p.pps.err, "PPSParser Errors"), multierror.Prefix(p.slice.err, "SliceParser Errors"), p.annexb.Error())
}
