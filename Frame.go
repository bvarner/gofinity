package gofinity

import (
	"github.com/npat-efault/crc16"
	"errors"
	"bytes"
	"encoding/binary"
)

// A Frame Header
type Header struct {
	Destination uint16
	Source      uint16
	Length      uint8
	reserved    uint16 // Not sure what these two bytes are yet
	Operation   uint8
}

// A Communication Frame (message)
type Frame struct {
	header   Header
	data     []byte
	checksum []byte
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

	headerDataLength := len(buf) - 2 // Length of Header + Data - Checksum

	// Calculate the checksum and compare...
	// TODO: See comment in checksum() function below.
	bufChecksum := checksum(buf[:headerDataLength])
	if !bytes.Equal(bufChecksum, buf[headerDataLength:]) {
		return nil, errors.New("Frame Checksum mismatch")
	}

	return &Frame{
		header: Header{
			Destination: binary.BigEndian.Uint16(buf[0:2]),
			Source:      binary.BigEndian.Uint16(buf[2:4]),
			Length:      buf[4],
			reserved:    binary.BigEndian.Uint16(buf[5:7]),
			Operation:   buf[7],
		},
		data:     buf[8:headerDataLength],
		checksum: buf[headerDataLength:],
	}, nil
}


// Global Checksum configuration.
var crcConfig = &crc16.Conf{
	Poly:   0x8005, BitRev: true,
	IniVal: 0x0, FinVal: 0x0,
	BigEnd: false,
}

// Unexported function to calculate checksums.
func checksum(b []byte) []byte {
	// TODO: Evaluate if we can cut this down to just a Checksum(crcConfig, b) and compare the resultant uint16.
	s := crc16.New(crcConfig)
	s.Write(b)
	return s.Sum(nil)
}

