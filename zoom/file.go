package zoom

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"go.uber.org/zap"
)

type File struct {
	file       *os.File
	size       int64
	currOffset int64
	log        *zap.SugaredLogger
}

func OpenFile(filename string, log *zap.SugaredLogger) (*File, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("Failed to open file %s: %w", filename, err)
	}
	fileSize, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, fmt.Errorf("Failed to seek to end: %w", err)
	}
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("Failed to seek to start: %w", err)
	}
	f := &File{
		file:       file,
		log:        log.Named("file"),
		size:       fileSize,
		currOffset: 0,
	}
	return f, nil
}

func (f *File) Close() {
	f.file.Close()
}

func (f *File) ReadExactBytes(number int) ([]byte, error) {
	bytes := make([]byte, number)

	num, err := f.file.Read(bytes)
	if err != nil {
		return nil, fmt.Errorf("Failed to read %d bytes: %w", number, err)
	}

	if num != number {
		return nil, fmt.Errorf("Failed to read requested amount: %d vs %d", num, number)
	}

	f.currOffset += int64(num)
	//f.log.Debugw("Read bytes from file", "num", num)

	return bytes, nil
}

func (f *File) ReadBytes(number int) ([]byte, error) {
	// round to nearest 4
	number = (number + 3) & 0xfffffffc
	bytes := make([]byte, number)

	num, err := f.file.Read(bytes)
	if err != nil {
		return nil, fmt.Errorf("Failed to read %d bytes: %w", number, err)
	}

	if num != number {
		return nil, fmt.Errorf("Failed to read requested amount: %d vs %d", num, number)
	}

	f.currOffset += int64(num)
	//f.log.Debugw("Read bytes from file", "num", num)

	return bytes, nil
}

func (f *File) ReadStruct(number int, struc interface{}) error {
	data, err := f.ReadBytes(number)
	if err != nil {
		return err
	}
	buffer := bytes.NewBuffer(data)
	err = binary.Read(buffer, binary.LittleEndian, struc)
	if err != nil {
		return fmt.Errorf("Failed to convert data: %v to %T: %w", data, struc, err)
	}

	return nil
}

func (f *File) ReadBeginning() error {
	head := &Header{}

	f.ReadStruct(HeaderSize, head)
	if head.Header != header {
		return fmt.Errorf("Failed to read file header, expected header %x, got %x", header, head.Header)
	}

	if head.Trailer != trailer {
		return fmt.Errorf("Failed to read file header, expected trailer %x, got %x", trailer, head.Trailer)
	}

	f.log.Infow("Read beginning of file", "initial_offset", head.FileOffset, "version_info", head.VersionInfo, "version_number", head.VersionNumber())

	act, err := f.file.Seek(int64(head.FileOffset), io.SeekStart)
	if err != nil {
		return fmt.Errorf("Failed to seek to %d: %w", head.FileOffset, err)
	}

	if act != int64(head.FileOffset) {
		return fmt.Errorf("Failed to seek to %d, only seeked to %d", head.FileOffset, act)
	}

	f.currOffset = int64(head.FileOffset)

	return nil
}

func (f *File) HasData() bool {
	return f.currOffset < f.size
}

type Packet interface {
	// Called by the file, before reading the trailer.
	// Allows a packet to read more data
	ReadPacket(f *File) error
}

func (f *File) ReadPacket(pkt Packet) error {
	headTrail := uint32(0)

	err := binary.Read(f.file, binary.LittleEndian, &headTrail)
	f.currOffset += 4
	if err != nil {
		return fmt.Errorf("Failed to read header: %w", err)
	}
	if headTrail != header {
		return fmt.Errorf("Failed to read header, expected %x, got %x", header, headTrail)
	}

	if err = pkt.ReadPacket(f); err != nil {
		return err
	}

	err = binary.Read(f.file, binary.LittleEndian, &headTrail)
	f.currOffset += 4
	if err != nil {
		return fmt.Errorf("Failed to read trail: %w", err)
	}

	if headTrail != trailer {
		return fmt.Errorf("Failed to read trailer, expected %x, got %x", trailer, headTrail)
	}

	return nil
}
