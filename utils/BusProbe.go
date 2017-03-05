package main

import (
	"flag"
	"fmt"
	"github.com/bvarner/gofinity"
	"os"
	"time"
	log "github.com/Sirupsen/logrus"
)

func main() {
	fmt.Println("Starting up BusProbe.")

	serialPort := flag.String("s", "", "path to serial port device")
	replayFile := flag.String("f", "", "binary capture file to replay")

	flag.Parse()

	var transceiver gofinity.Transceiver = nil

	if len(*serialPort) != 0 {
		transceiver = gofinity.NewSerialTransceiver(*serialPort)
	}
	if len(*replayFile) != 0 {
		transceiver = gofinity.NewFileBusReplayer(*replayFile)
	}

	if transceiver == nil {
		fmt.Println("You must specify either -s (serial device) or -f (replay flie)")
		flag.PrintDefaults()
		os.Exit(1)
	}

	err := transceiver.Open()
	if err != nil {
		log.Fatal(err)
	}

	bus := gofinity.NewBus(transceiver)
	bus.Probe(func(frame *gofinity.Frame) {
		log.Info(frame)
	})

	err = bus.Start()
	if err != nil {
		log.Fatal(err)
	}
	defer bus.Shutdown()

	for transceiver.IsOpen() {
		fmt.Println("Sleeping 5 seconds for transceiver isOpen test...")
		time.Sleep(time.Second * 5)
	}

	fmt.Println("Transceiver closed")
}
