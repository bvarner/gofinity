package gofinity

import (
	"github.com/npat-efault/crc16"
	"errors"
	"encoding/binary"
	log "github.com/Sirupsen/logrus"
	"fmt"
	"strconv"
)

var Operations = map[uint8]string{
	0x06: "ACK",
	0x0b: "READ_REQUEST",
	0x0c: "WRITE_REQUEST",
	0x15: "ERROR",
}

// A Frame Header
type Header struct {
	Destination uint16
	Source      uint16
	Length      uint8
	reserved1   uint8 // Not sure what these two bytes are yet
	reserved2   uint8
	Operation   uint8
}

// A Communication Frame (message)
type Frame struct {
	header   Header
	data     []byte
	checksum uint16
}

// Decodes and creates a new Frame from the given buffer.
func NewFrame(buf []byte) (*Frame, error) {
	empty := true
	for _, c := range buf {
		if c != 0 {
			empty = false
			break
		}
	}

	if empty {
		return nil, errors.New("No Frame Content")
	}

	// End of the buffer is a 2 byte checksum.
	headerDataLength := len(buf) - 2 // Length of Header + Data - Checksum

	// Calculate a checksum from the bytes we've received.
	rxChecksum := crc16.Checksum(crcConfig, buf[:headerDataLength])

	// Read the checksum from the buffer
	txChecksum := binary.LittleEndian.Uint16(buf[headerDataLength:])

	if rxChecksum != txChecksum {
		log.Error(fmt.Sprintf("Checksum Mismatch: %x != %x", rxChecksum, txChecksum))
		return nil, errors.New("Frame Checksum mismatch")
	}

	// Checksum matches. Construct the frame.
	return &Frame{
		header: Header{
			Destination: binary.LittleEndian.Uint16(buf[0:2]),
			Source:      binary.LittleEndian.Uint16(buf[2:4]),
			Length:      buf[4], // uint8
			reserved1:   buf[5], // Not sure what this byte and the next are.
			reserved2:   buf[6],
			Operation:   buf[7], // uint8
		},
		data:     buf[8:headerDataLength],
		checksum: txChecksum, // uint16
	}, nil
}

func (header *Header) String() string {
	return fmt.Sprintf("%4x -> %4x [%3d] : %s %s : %14s",
		header.Source, header.Destination, header.Length,
		strconv.FormatUint(uint64(header.reserved1), 2),
		strconv.FormatUint(uint64(header.reserved2), 2),
		Operations[header.Operation])

}

func (f *Frame) String() string {
	return fmt.Sprintf("%s : %x", f.header.String(), f.data)
}

// Global Checksum configuration.
var crcConfig = &crc16.Conf{
	Poly:   0x8005, BitRev: true,
	IniVal: 0x0, FinVal: 0x0,
	BigEnd: false,
}
