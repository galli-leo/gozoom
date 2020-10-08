package zoom

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	header  uint32 = 0x2C05F158
	trailer        = 0x84AD52E2
)

const (
	HeaderSize int = 96
	CmdSize    int = 64
	SampleSize int = 48
)

type Header struct {
	Header  uint32
	Trailer uint32
	// + 8
	_ [8]byte // Unknown
	_ [8]byte // Unknown but has something to do with file size?
	_ [8]byte // Unknown
	// + 32
	VersionInfo uint32
	FileOffset  uint32   // offset to start of actual data
	_           [56]byte // Unknown and unused at the moment.
}

func (h *Header) IsMC() bool {
	return h.VersionNumber() > 1
}

func (h *Header) VersionNumber() int {
	if h.VersionInfo>>16 < 0xE {
		if h.VersionInfo > 0x4E20 {
			if h.VersionInfo > 0x589D {
				if h.VersionInfo > 0xB5F8 {
					if h.VersionInfo > 0xBBC7 {
						if h.VersionInfo > 0xC601 {
							return 5
						}
						return 4
					}
					return 3
				}
				return 2
			}
			return 1
		}
		return 0
	}
	return -1
}

type CmdHeader struct {
	SizeToRead     uint32
	Type           int32
	_              [8]byte // Unknown
	NameIdent      uint32
	_              [4]byte // Unknown
	TimingA        uint32
	_              [22]byte // Unknown
	SomeType       uint16
	_              [4]byte
	AdditionalSize int32
	_              [4]byte // Unknown
}

type CmdPacket struct {
	*CmdHeader
	AdditionalData []byte
}

func NewCmdPacket() *CmdPacket {
	return &CmdPacket{
		CmdHeader: &CmdHeader{},
	}
}

func (p *CmdPacket) String() string {
	addData := ""
	if p.AdditionalSize > 0 {
		addData = string(p.AdditionalData)
	}
	return fmt.Sprintf("<Cmd (%d): Person 0x%x, Timing: 0x%x, Type: 0x%x: %s>", p.Type, p.NameIdent, p.TimingA, p.SomeType, addData)
}

func (p *CmdPacket) ReadPacket(f *File) error {
	if err := f.ReadStruct(CmdSize, p.CmdHeader); err != nil {
		return fmt.Errorf("Failed to read CmdHeader: %w", err)
	}
	if p.AdditionalSize > 0 {
		data, err := f.ReadExactBytes(int(p.CmdHeader.AdditionalSize))
		if err != nil {
			return fmt.Errorf("Failed to read %d bytes of additional data: %w", p.AdditionalSize, err)
		}

		p.AdditionalData = data
	}
	return nil
}

type MediaType int32

const (
	Audio            MediaType = 4
	VideoWebCam      MediaType = 0x10
	VideoScreenShare MediaType = 0x20
	Avatar           MediaType = 0x40
	Cursor           MediaType = 0x1000
)

type SampleHeader struct {
	Type         int32
	_            [4]byte // + 8
	TimingA      int64   // + 16
	TimingB      int64   // + 24
	_            [8]byte // + 32
	DataSize     int32
	PropertySize int32 // + 40
	_            [8]byte
}

type SamplePacket struct {
	*SampleHeader
	Data     []byte
	Property []byte
}

func NewSamplePacket() *SamplePacket {
	return &SamplePacket{
		SampleHeader: &SampleHeader{},
	}
}

func (p *SamplePacket) ReadPacket(f *File) error {
	if err := f.ReadStruct(SampleSize, p.SampleHeader); err != nil {
		return fmt.Errorf("Failed to read sample packet: %w", err)
	}

	if p.Type < 0 {
		return nil
	}

	if p.PropertySize > 0 {
		data, err := f.ReadBytes(int(p.PropertySize))
		if err != nil {
			return fmt.Errorf("Failed to read property: %w", err)
		}
		p.Property = data
	}

	if p.DataSize > 0 {
		data, err := f.ReadBytes(int(p.DataSize))
		if err != nil {
			return fmt.Errorf("Failed to read data: %w", err)
		}
		p.Data = data[:p.DataSize]
	}

	// f.log.Debugf("Read Packet: %v, %d, %d", p.SampleHeader, p.DataSize, p.PropertySize)

	return nil
}

func (p *SamplePacket) MediaType() MediaType {
	return MediaType(p.Type)
}

type VideoProp struct {
	NameIdent int32
	_         [4]byte
	Width     int32
	Height    int32
	_         [8]byte
}

func (p *SamplePacket) VideoProp() *VideoProp {
	if p.MediaType() == VideoScreenShare {
		if p.PropertySize > 8 {
			vid := &VideoProp{}

			buf := bytes.NewBuffer(p.Property)
			binary.Read(buf, binary.LittleEndian, vid)
			return vid
		}
	}

	return nil
}

type CursorProp struct {
	X         int32
	Y         int32
	Width     int32
	Height    int32
	NameIdent int32
	Type      int32
}

func (p *SamplePacket) CursorProp() *CursorProp {
	if p.MediaType() == Cursor {
		if p.PropertySize > 8 {
			cur := &CursorProp{}

			buf := bytes.NewBuffer(p.Property)
			binary.Read(buf, binary.LittleEndian, cur)
			return cur
		}
	}

	return nil
}
