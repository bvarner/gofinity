package gofinity

import (
	"io"
	"github.com/tarm/serial"
	"os"
	"time"
	log "github.com/Sirupsen/logrus"
)

// BusTransceiver abstracts a ReadWriteCloser and the functions expected by BusNode for opening / closing streams.
type BusTransceiver interface {
	io.ReadWriteCloser
	Open() error
	IsOpen() bool
	Valid() bool
}

// SerialBusTransceiver is a BusTransceiver that operations on serial ports. Surprise!
type SerialBusTransceiver struct {
	device string
	port   *serial.Port
}

func NewSerialBusTransceiver(device string) (*SerialBusTransceiver) {
	return &SerialBusTransceiver{device: device}
}

func (st *SerialBusTransceiver) Read(p []byte) (n int, err error) {
	return st.port.Read(p)
}

func (st *SerialBusTransceiver) Write(p []byte) (n int, err error) {
	return st.port.Write(p)
}

func (st *SerialBusTransceiver) Close() error {
	err := st.port.Close()
	st.port = nil
	return err
}

func (st *SerialBusTransceiver) Open() error {
	var err error
	config := &serial.Config{Name: st.device, Baud: 38400, ReadTimeout: (time.Second * 30)}
	st.port, err = serial.OpenPort(config)
	if err != nil {
		st.port = nil
	}
	return err
}

func (st *SerialBusTransceiver) IsOpen() bool {
	return st.port != nil
}

func (st *SerialBusTransceiver) Valid() bool {
	return true
}

// FileBusReplayer is a BusTransceiver that turns writes into 'no-ops', but allows for probing previously recorded bus logs.
// if a FileBusReplayer hits an EOF, it's considered no longer valid.
type FileBusReplayer struct {
	fileName string
	file     *os.File
	atEOF    bool
}

func NewFileBusReplayer(file string) (*FileBusReplayer) {
	return &FileBusReplayer{fileName: file, file: nil, atEOF: false}
}

func (fb *FileBusReplayer) Read(p []byte) (n int, err error) {
	n, err = fb.file.Read(p)
	if err == io.EOF {
		fb.atEOF = true
	}
	return n,err
}

func (fb *FileBusReplayer) Write(p []byte) (n int, err error) {
	// Act like we wrote the whole thing
	return len(p), nil
}

func (fb *FileBusReplayer) Close() error {
	err := fb.file.Close()
	fb.file = nil
	return err
}

func (fb *FileBusReplayer) Open() error {
	var err error
	log.Info("Attempting to Open %s", fb.fileName)
	fb.file, err = os.Open(fb.fileName)
	if err != nil {
		fb.file = nil
	}
	return err
}

func (fb *FileBusReplayer) IsOpen() bool {
	return fb.file != nil
}

func (fb *FileBusReplayer) Valid() bool {
	return !fb.atEOF
}
