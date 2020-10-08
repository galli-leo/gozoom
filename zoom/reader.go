package zoom

import (
	"io"

	"go.uber.org/zap"
)

func NewSampleReader(log *zap.SugaredLogger) *SampleReader {
	return &SampleReader{log: log.Named("SampleReader")}
}

type SampleReader struct {
	f    *File
	log  *zap.SugaredLogger
	curr *SamplePacket
}

func (s *SampleReader) Open(filename string) error {
	var err error
	s.f, err = OpenFile(filename, s.log)
	if err == nil {
		err = s.f.ReadBeginning()
	}
	return err
}

func (s *SampleReader) Close() {
	s.f.Close()
}

func (s *SampleReader) Next() bool {
	if s.f.HasData() {
		s.curr = NewSamplePacket()
		err := s.f.ReadPacket(s.curr)
		if err != nil {
			s.log.Errorf("Failed to read packet: %v", err)
		}
		return err == nil
	}

	return false
}

func (s *SampleReader) Current() *SamplePacket {
	return s.curr
}

type SampleDataFilter func(*SamplePacket) bool

type SampleDataTransformer func([]byte) []byte

func SampleDataNoTransform(b []byte) []byte {
	return b
}

func NewSampleDataReader(r *SampleReader, filter SampleDataFilter, trans SampleDataTransformer, log *zap.SugaredLogger) *SampleDataReader {
	return &SampleDataReader{
		reader: r,
		log:    log.Named("SampleDataReader"),
		filter: filter,
		trans:  trans,
		buf:    []byte{},
	}
}

type SampleDataReader struct {
	reader *SampleReader
	log    *zap.SugaredLogger
	filter SampleDataFilter
	trans  SampleDataTransformer
	buf    []byte
}

func (r *SampleDataReader) readPacketData() error {
	for r.reader.Next() {
		pkt := r.reader.Current()
		if r.filter(pkt) {
			if len(pkt.Data) > 0 {
				r.buf = r.trans(pkt.Data)
				return nil
			}
		}
	}

	return io.EOF
}

func (r *SampleDataReader) copyTo(b []byte) (n int, err error) {
	if len(r.buf) == 0 {
		if err = r.readPacketData(); err != nil {
			return
		}
	}

	n = copy(b, r.buf)
	r.buf = r.buf[n:]

	return
}

func (r *SampleDataReader) Read(p []byte) (n int, err error) {
	left := len(p)
	var diff int
	for left > 0 {
		diff, err = r.copyTo(p[n:])
		if err != nil {
			return
		}
		left -= diff
		n += diff
	}

	return
}
