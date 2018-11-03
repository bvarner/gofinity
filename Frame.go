package gofinity

import (
	"github.com/npat-efault/crc16"
	"errors"
	"encoding/binary"
	log "github.com/Sirupsen/logrus"
	"fmt"
	"strconv"
)

const ACK02 = 0x02
const ACK06 = 0x06
const READ_TABLE_BLOCK = 0x0b
const WRITE_TABLE_BLOCK = 0x0c
const CHANGE_TABLE_NAME = 0x10
const NACK = 0x15
const ALARM_PACKET = 0x1e
const READ_OBJECT_DATA = 0x22
const READ_VARIABLE = 0x62
const WRITE_VARIABLE = 0x63
const AUTO_VARIABLE = 0x64
const READ_LIST = 0x75

var Operations = map[uint8]string{
	ACK02:             "ACK02",
	ACK06:             "ACK06",
	READ_TABLE_BLOCK:  "READ",
	WRITE_TABLE_BLOCK: "WRITE",
	CHANGE_TABLE_NAME: "CHGTBN",
	NACK:              "NACK",
	ALARM_PACKET:      "ALARM",
	READ_OBJECT_DATA:  "OBJRD",
	READ_VARIABLE:     "RDVAR",
	WRITE_VARIABLE:    "FORCE",
	AUTO_VARIABLE:     "AUTO",
	READ_LIST:         "LIST",
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
	Header   Header
	payload  []byte
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
		Header: Header{
			Destination: binary.LittleEndian.Uint16(buf[0:2]),
			Source:      binary.LittleEndian.Uint16(buf[2:4]),
			Length:      buf[4], // uint8
			reserved1:   buf[5], // Not sure what this byte and the next are.
			reserved2:   buf[6],
			Operation:   buf[7], // uint8
		},
		payload:  buf[8:headerDataLength],
		checksum: txChecksum, // uint16
	}, nil
}

func NewProbeDeviceFrame(source uint16, destination uint16, exportIdx uint16, offset uint8) (*Frame) {
	probeFrame := Frame{
		Header: Header{
			Destination: destination,
			Source:      source,
			Operation:   READ_TABLE_BLOCK,
		},
	}

	// Put together the payload.
	probeFrame.payload = make([]byte, 3)

	// Export index, first entry.
	binary.BigEndian.PutUint16(probeFrame.payload, exportIdx)
	probeFrame.payload[2] = offset

	return &probeFrame
}

func (header *Header) String() string {
	return fmt.Sprintf("%4x -> %4x [%3d] : %s %s : %14s",
		header.Source, header.Destination, header.Length,
		strconv.FormatUint(uint64(header.reserved1), 2),
		strconv.FormatUint(uint64(header.reserved2), 2),
		Operations[header.Operation])

}

func (frame *Frame) Encode() ([]byte) {
	frame.Header.Length = uint8(len(frame.payload))

	// Create a buffer big enough.
	buf := make([]byte, frame.Header.Length+8) // Header length in bytes.

	binary.LittleEndian.PutUint16(buf[0:2], frame.Header.Destination)
	binary.LittleEndian.PutUint16(buf[2:4], frame.Header.Source)
	buf[4] = frame.Header.Length
	buf[5] = 0x00
	buf[6] = 0x00
	buf[7] = frame.Header.Operation
	copy(buf[8:], frame.payload)

	// Calculate the CRC.
	crc := crc16.Checksum(crcConfig, buf)

	// Make a new buffer big enough to hold the output + crc
	outbuf := make([]byte, len(buf)+2)
	copy(outbuf, buf)
	binary.LittleEndian.PutUint16(outbuf[len(buf):], crc)

	return outbuf
}

func (frame *Frame) String() string {
	return fmt.Sprintf("%s : %s", frame.Header.String(), frame.Payload())
}

func (frame *Frame) Payload() string {
	if len(frame.payload) == 1 && frame.payload[0] == 0x00 {
		return ""
	} else {
		expStruct := binary.BigEndian.Uint16(frame.payload[0:2]) // Index of array of pointers, pointing to memory addresses of structures containing data type, name, and count of similar in-memory structs.
		expIdx := frame.payload[2]                               // Resolve Pointer to struct. Now, what index are we interested in?

		log.Debug(fmt.Sprintf("%x", frame.payload))

		return fmt.Sprintf("exports[%d].data[%d]", expStruct, expIdx)
	}
}

// Global Checksum configuration.
var crcConfig = &crc16.Conf{
	Poly:   0x8005, BitRev: true,
	IniVal: 0x0, FinVal: 0x0,
	BigEnd: false,
}
