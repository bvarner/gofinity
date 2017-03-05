package gofinity

import (
	"errors"
	"sync"
	log "github.com/Sirupsen/logrus"
)

// Defines a callback for Bus Probing.
type OnFrameReceived func(*Frame)

// Constants for Bus.status
const (
	READY    = uint8(0x00)
	RUNNING  = uint8(0x01)
	STOPPING = uint8(0x02)
	INVALID  = uint8(0x03)
)

// Bus defines frame-based interactions atop a Transceiver
type Bus struct {
	waitGroup   sync.WaitGroup
	transceiver Transceiver
	probes      []OnFrameReceived
	status      uint8
}

// Constructs a new Bus for a BusTransciever
func NewBus(transceiver Transceiver) (*Bus) {
	busNode := &Bus{
		transceiver: transceiver,
		probes:      []OnFrameReceived{},
		waitGroup:   sync.WaitGroup{},
		status:      READY,
	}

	return busNode
}

func (bus *Bus) Probe(received OnFrameReceived) {
	bus.probes = append(bus.probes, received)
}

// Internal read loop for reading Frames from the transceiver.
func (bus *Bus) readLoop() {
	log.Info("Starting Bus.readLoop()")
	defer log.Info("Bus.readLoop() finished")
	defer bus.waitGroup.Done()

	frameBuf := []byte{}
	readBuf := make([]byte, 1024)

	// If the transceiver is no longer valid, bail.
	for bus.status == RUNNING && bus.transceiver.Valid() {
		// If the transceiver isn't open, reset the framebuffer and (re)open.
		if !bus.transceiver.IsOpen() {
			frameBuf = []byte{}
			bus.transceiver.Open()
		}

		// Try to read some bytes.
		n, readErr := bus.transceiver.Read(readBuf)

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
				for _, probe := range bus.probes {
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
			bus.transceiver.Close()
		}
	}

	if !bus.transceiver.Valid() {
		bus.status = INVALID
		log.Info("Transceiver no longer valid for reading.")
	}
}

// Starts the Bus's I/O loops.
func (bus *Bus) Start() error {
	if bus.status != READY {
		return errors.New("Bus not in 'READY' state.")
	}
	bus.status = RUNNING

	bus.waitGroup.Add(1)
	go bus.readLoop()

	return nil
}

// Stops the Bus's I/O loops.
// Blocks until all threads running for I/O terminate.
func (bus *Bus) Shutdown() error {
	if bus.status != RUNNING || bus.status != INVALID {
		return errors.New("Bus not 'RUNNING' or 'INVALID'.")
	}
	bus.status = STOPPING
	bus.waitGroup.Wait()
	bus.status = READY

	return nil
}
