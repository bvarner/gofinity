package gofinity

import (
	"io"
	"github.com/tarm/serial"
	"os"
	"time"
)

// BusTransceiver abstracts a ReadWriteCloser and the functions expected by BusNode for opening / closing streams.
type BusTransceiver interface {
	io.ReadWriteCloser
	Open() error
	IsOpen() bool
}

// SerialBusTransceiver is a BusTransceiver that operations on serial ports. Surprise!
type SerialBusTransceiver struct {
	device string
	port   *serial.Port
}

func NewSerialBusTransceiver(device string) (*SerialBusTransceiver) {
	return &SerialBusTransceiver{device: device}
}

// FileBusReplayer is a BusTransceiver that turns writes into 'no-ops', but allows for probing previously recorded bus logs.
type FileBusReplayer struct {
	fileName string
	file     *os.File
}

func NewFileBusReplayer(file string) (*FileBusReplayer) {
	return &FileBusReplayer{fileName: file}
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


func (fb *FileBusReplayer) Read(p []byte) (n int, err error) {
	return fb.file.Read(p)
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
	fb.file, err = os.OpenFile(fb.fileName, os.O_RDWR, 0666)
	_, err = fb.file.Seek(0, io.SeekStart)
	if err != nil {
		fb.file = nil
	}
	return err
}

func (fb *FileBusReplayer) IsOpen() bool {
	return fb.file != nil
}
