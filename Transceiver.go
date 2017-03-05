package gofinity

import (
	"io"
	"github.com/tarm/serial"
	"os"
	"time"
	log "github.com/Sirupsen/logrus"
)

// Transceiver abstracts a ReadWriteCloser and the functions expected by Bus for opening / closing streams.
type Transceiver interface {
	io.ReadWriteCloser
	Open() error
	IsOpen() bool
	Valid() bool
}

// SerialTransceiver is a Transceiver that operations on serial ports. Surprise!
type SerialTransceiver struct {
	device string
	port   *serial.Port
}

func NewSerialTransceiver(device string) (*SerialTransceiver) {
	return &SerialTransceiver{device: device}
}

func (st *SerialTransceiver) Read(p []byte) (n int, err error) {
	return st.port.Read(p)
}

func (st *SerialTransceiver) Write(p []byte) (n int, err error) {
	return st.port.Write(p)
}

func (st *SerialTransceiver) Close() error {
	err := st.port.Close()
	st.port = nil
	return err
}

func (st *SerialTransceiver) Open() error {
	var err error
	config := &serial.Config{Name: st.device, Baud: 38400, ReadTimeout: (time.Second * 30)}
	st.port, err = serial.OpenPort(config)
	if err != nil {
		st.port = nil
	}
	return err
}

func (st *SerialTransceiver) IsOpen() bool {
	return st.port != nil
}

func (st *SerialTransceiver) Valid() bool {
	return true
}

// FileReplayer is a Transceiver that turns writes into 'no-ops', but allows for probing previously recorded bus logs.
// if a FileReplayer hits an EOF, it's considered no longer valid.
type FileReplayer struct {
	fileName string
	file     *os.File
	atEOF    bool
}

func NewFileBusReplayer(file string) (*FileReplayer) {
	return &FileReplayer{fileName: file, file: nil, atEOF: false}
}

func (fb *FileReplayer) Read(p []byte) (n int, err error) {
	n, err = fb.file.Read(p)
	if err == io.EOF {
		fb.atEOF = true
	}
	return n,err
}

func (fb *FileReplayer) Write(p []byte) (n int, err error) {
	// Act like we wrote the whole thing
	return len(p), nil
}

func (fb *FileReplayer) Close() error {
	err := fb.file.Close()
	fb.file = nil
	return err
}

func (fb *FileReplayer) Open() error {
	var err error
	log.Info("Attempting to Open %s", fb.fileName)
	fb.file, err = os.Open(fb.fileName)
	if err != nil {
		fb.file = nil
	}
	return err
}

func (fb *FileReplayer) IsOpen() bool {
	return fb.file != nil
}

func (fb *FileReplayer) Valid() bool {
	return !fb.atEOF
}
