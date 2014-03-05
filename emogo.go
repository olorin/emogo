/*
The emogo package provides go bindings for emokit
(https://github.com/openyou/emokit). 
*/
package emogo

import (
	"errors"
	"unsafe"
)

// #include <emokit/emokit.h>
// #include <stdint.h>
// #cgo LDFLAGS: -lemokit
import "C"

// These are defined in emokit.h and reproduced here as cgo isn't
// linking them in for some reason.
const (
	EMOKIT_VID int = 0x21a1
	EMOKIT_PID  = 0x0001
	EmokitPacketSize = 32
)

type HeadsetType uint

const (
	DeveloperHeadset HeadsetType = 0
	ConsumerHeadset HeadsetType = 1
)

// EmokitContext represents a connection to an EPOC device. 
type EmokitContext struct {
	eeg *C.struct_emokit_device
}

func NewEmokitContext(t HeadsetType) (*EmokitContext, error) {
	e := new(EmokitContext)
	e.eeg = C.emokit_create()
	ret := C.emokit_open(e.eeg, C.int(EMOKIT_VID), C.int(EMOKIT_PID), C.uint(t))
	if ret != 0 {
		return nil, errors.New("Cannot access device.")
	}
	return e, nil
}

type EmokitFrame struct {
	raw []byte
	rendered C.struct_emokit_frame
}

func NewEmokitFrame() *EmokitFrame {
	f := new(EmokitFrame)
	f.raw = make([]byte, EmokitPacketSize)
	return f
}

// readData reads data from the EPOC dongle and returns 0 on success, <0
// on error.
func (e *EmokitContext) readData() error {
	n := C.emokit_read_data(e.eeg)
	if n >= 0 {
		return nil
	}
	return errors.New("emokit_read_data failed")
}

func (e *EmokitContext) getNextFrame() (*EmokitFrame, error) {
	f := NewEmokitFrame()
	f.rendered = C.emokit_get_next_frame(e.eeg)
	if f.rendered.counter == 0 {
		return nil, errors.New("Could not read raw packet.")
	}
	C.emokit_get_raw_frame(e.eeg, (*C.uchar)(unsafe.Pointer(&f.raw[0])))
	return f, nil
}

// GetFrame returns the next available EPOC frame. If there is no frame
// to be read, the error value will be EAGAIN.
func (e *EmokitContext) GetFrame() (*EmokitFrame, error) {
	err := e.readData()
	if err == nil {
		f, err := e.getNextFrame()
		if err != nil {
			return nil, err
		}
		return f, nil
	}
	return nil, err
}

func (e *EmokitContext) Count() int {
	n := C.emokit_get_count(e.eeg, C.int(EMOKIT_VID), C.int(EMOKIT_PID))
	return int(n)
}

// Raw returns the (unencrypted) raw EPOC frame.
func (f *EmokitFrame) Raw() []byte {
	return f.raw
}

// Gyro returns the current (x,y) of the frame's Gyro value.
func (f *EmokitFrame) Gyro() (int,int) {
	return int(f.rendered.gyroX), int(f.rendered.gyroY)
}
