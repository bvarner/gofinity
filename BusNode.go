package gofinity

import (
	"errors"
	"sync"
	log "github.com/Sirupsen/logrus"
)

// Defines a callback for Bus Probing.
type OnFrameReceived func(*Frame)

// Constants for BusNode.status
const (
	READY    = uint8(0x00)
	RUNNING  = uint8(0x01)
	STOPPING = uint8(0x02)
	INVALID  = uint8(0x03)
)

// BusNode defines frame-based interactions atop a BusTransceiver
type BusNode struct {
	waitGroup   sync.WaitGroup
	transceiver BusTransceiver
	probes      []OnFrameReceived
	status      uint8
}

// Constructs a new BusNode for a BusTransciever
func NewBusNode(transceiver BusTransceiver) (*BusNode) {
	busNode := &BusNode{
		transceiver: transceiver,
		probes:      []OnFrameReceived{},
		waitGroup:   sync.WaitGroup{},
		status:      READY,
	}

	return busNode
}

func (busNode *BusNode) Probe(received OnFrameReceived) {
	busNode.probes = append(busNode.probes, received)
}

// Internal read loop for reading Frames from the transceiver.
func (busNode *BusNode) readLoop() {
	log.Info("Starting BusNode.readLoop()")
	defer log.Info("BusNode.readLoop() finished")
	defer busNode.waitGroup.Done()

	frameBuf := []byte{}
	readBuf := make([]byte, 256)

	// If the transceiver is no longer valid, bail.
	for busNode.status == RUNNING && busNode.transceiver.Valid() {
		// If the transceiver isn't open, reset the framebuffer and (re)open.
		if !busNode.transceiver.IsOpen() {
			frameBuf = []byte{}
			busNode.transceiver.Open()
		}

		// Try to read some bytes.
		n, readErr := busNode.transceiver.Read(readBuf)

		// Append to the frame buffer the bytes we just read.
		frameBuf = append(frameBuf, readBuf[:n]...)

		for {
			// Make sure we have at least a full header.
			if len(frameBuf) < 10 {
				break
			}

			// Byte 5 of valid frames tell us how long the frame is, plus header length.
			frameLength := int(frameBuf[4]) + 10
			if len(frameBuf) < frameLength {
				break;
			}

			frameSlice := frameBuf[:frameLength]

			frame, err := NewFrame(frameSlice)
			if err == nil {
				for _, probe := range busNode.probes {
					probe(frame)
				}

				// This portion, (0 - length), handled.
				// Slice (advance) the buffer.
				frameBuf = frameBuf[:copy(frameBuf, frameBuf[frameLength:])]
			} else {
				// Corrupt Message, or not quite a frame yet.
				// Advance one byte, try again.
				frameBuf = frameBuf[:copy(frameBuf, frameBuf[1:])]
			}
		}

		if readErr != nil {
			log.Warn("Erorr reading : ", readErr)
			busNode.transceiver.Close()
		}
	}

	if !busNode.transceiver.Valid() {
		busNode.status = INVALID
		log.Info("Transceiver no longer valid for reading.")
	}
}

// Starts the BusNode's I/O loops.
func (busNode *BusNode) Start() error {
	if busNode.status != READY {
		return errors.New("BusNode not in 'READY' state.")
	}
	busNode.status = RUNNING

	busNode.waitGroup.Add(1)
	go busNode.readLoop()

	return nil
}

// Stops the BusNode's I/O loops.
// Blocks until all threads running for I/O terminate.
func (busNode *BusNode) Shutdown() error {
	if busNode.status != RUNNING {
		return errors.New("BusNode not 'RUNNING'.")
	}
	busNode.status = STOPPING
	busNode.waitGroup.Wait()
	busNode.status = READY

	return nil
}
