package main

import (
	"flag"
	"github.com/bvarner/gofinity"
	log "github.com/Sirupsen/logrus"
	"time"
	"fmt"
	"os"
)

// This is a really nasty, mean little tool, that disrespects "proper" device interaction on the bus by writing whenever
// it darn well wants to (which is really not the way this "should" be working, methinks)
//
// I have yet to record enough logs including device adding to the network (i.e. I haven't power-cycled my heat pump
// while logging everything yet)
//
// But at least this exercises some basic frame encoding and writing, and it seems to work reasonably well.
func main() {
	log.SetLevel(log.DebugLevel)

	// I'm using terminology that's not "in line" with what everyone else is using.
	// Firstly, I reject the term 'table' on principle.
	// My current hypothesis is that these are pointers to pointers to structs (well, that's the simplified version)
	// or at the very least array indexes inside an array -- likely arrays of pointers to structs. Yep.
	// I'd kill for a copy of the header file that defines the standards for this protocol.
	// I also bet there's a few engineers at Carrier that laugh about the current attempts to decipher this stuff.
	serialPort := flag.String("s", "", "path to serial port device")
	devAddr := flag.Uint("a", 0x121, "Address of this device")
	probeAddr := flag.Uint("p", 0xf1f1, "Address of the device to probe")
	exportIdx := flag.Uint( "i", 0x0001, "Address (index) of exported API to probe")
	offset := flag.Uint("o", 0x01, "Offset within the API")

	flag.Parse()

	var transceiver gofinity.Transceiver = nil

	// We're only going to let this tool work on real serial connections.
	if len(*serialPort) != 0 {
		transceiver = gofinity.NewSerialTransceiver(*serialPort)
	}

	if transceiver == nil {
		defer flag.PrintDefaults()
		log.Fatal("You must specify a -s (serial device) for device probing.")
	}

	// Open the transciever (serial port)
	err := transceiver.Open()
	if err != nil {
		log.Fatal(err)
	}

	// Create a new Bus instance on the transceiver, and add a probe for incoming Frames.
	bus := gofinity.NewBus(transceiver)
	bus.Probe(func(frame *gofinity.Frame) {
		// If it's sent to the destination we're listening on...
		if frame.Header.Destination == uint16(*devAddr) {
			log.Info(frame)
			bus.Shutdown()
		}
	})

	// Start up the bus interaction.
	err = bus.Start()
	if err != nil {
		log.Fatal(err)
	}
	defer bus.Shutdown()


	// TODO: In the future, this would be queued for a write in the bus, where we'd be issuing our outgoing frames
	// after the coordinator polls us.
	// Or, if we're a coordinator, after a few hundred ms of inactivity on the bus.
	// For now, create a frame and slam that sucker into the bus, consequences be darned.
	probeFrame := gofinity.NewProbeDeviceFrame(uint16(*devAddr), uint16(*probeAddr), uint16(*exportIdx), uint8(*offset))
	toSend := probeFrame.Encode()

	if transceiver.IsOpen() {
		log.Debug(fmt.Sprintf("Sending %s", probeFrame.String()))
		log.Debug("Writing ", len(toSend), " bytes")
		n, error := transceiver.Write(toSend)

		log.Info("Wrote ", n, " bytes with error: ", error)
	}


	// FWIW, we don't always get a response back in the timeframe you want it in.
	for transceiver.IsOpen() {
		log.Debug("Sleeping 5 seconds for transceiver isOpen test...")
		time.Sleep(time.Second * 5)
	}

	log.Info("Transceiver closed")

}
