package github.com/bvarner/gofinity

import (
	"errors"
	"sync"
	log "github.com/Sirupsen/logrus"
)

// Defines a callback for Bus Probing.
type onFrameReceived func(*Frame)

// Constants for BusNode.status
const (
	READY    = uint8(0x00)
	RUNNING  = uint8(0x01)
	STOPPING = uint8(0x02)
)

// BusNode defines frame-based interactions atop a BusTransceiver
type BusNode struct {
	waitGroup   sync.WaitGroup
	transceiver BusTransceiver
	probes      []onFrameReceived
	status      uint8
}

// Constructs a new BusNode for a BusTransciever
func NewBusNode(transceiver BusTransceiver) (*BusNode) {
	busNode := &BusNode{
		transceiver: transceiver,
		probes:      []onFrameReceived{},
		waitGroup:   sync.WaitGroup{},
		status:      READY,
	}

	return busNode
}

// Internal read loop for reading Frames from the transceiver.
func (busNode *BusNode) readLoop() {
	log.Info("Starting BusNode.readLoop()")
	defer log.Info("BusNode.readLoop() finished")
	defer busNode.waitGroup.Done()

	frameBuf := []byte{}
	readBuf := make([]byte, 1024)

	for busNode.status == RUNNING {
		// If the transceiver isn't open, reset the framebuffer and (re)open.
		if !busNode.transceiver.IsOpen() {
			frameBuf = []byte{}
			busNode.transceiver.Open()
		}

		// Try to read some bytes.
		n, err := busNode.transceiver.Read(readBuf)
		if n == 0 || err != nil {
			log.Printf("error reading from %s: %s", busNode.transceiver, err.Error())
			if busNode.transceiver.IsOpen() {
				busNode.transceiver.Close()
			}
			// Next trip through the loop will re-open the stream.
			continue
		}

		// Append to the frame buffer the bytes we just read.
		frameBuf = append(frameBuf, readBuf[:n]...)

		//
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
