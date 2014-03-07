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

// Shutdown closes the connection to the EPOC and frees associated 
// memory.
func (c *EmokitContext) Shutdown() {
	C.emokit_close(c.eeg)
	C.emokit_delete(c.eeg)
}

type EmokitSensor struct {
	value int
	quality int
}

type EmokitFrame struct {
	raw []byte
	rendered C.struct_emokit_frame
	F3 EmokitSensor
	FC6 EmokitSensor
	P7 EmokitSensor
	T8 EmokitSensor
	F7 EmokitSensor
	F8 EmokitSensor
	T7 EmokitSensor
	P8 EmokitSensor
	AF4 EmokitSensor
	F4 EmokitSensor
	AF3 EmokitSensor
	O2 EmokitSensor
	O1 EmokitSensor
	FC5 EmokitSensor
}

func NewEmokitFrame() *EmokitFrame {
	f := new(EmokitFrame)
	f.raw = make([]byte, EmokitPacketSize)
	return f
}

// parseSensors populates the EmokitSensor elements of the frame from
// the C struct.
func (f *EmokitFrame) parseSensors() {
	f.F3.value = int(f.rendered.F3)
	f.F3.quality = int(f.rendered.cq.F3)
	f.FC6.value = int(f.rendered.FC6)
	f.FC6.quality = int(f.rendered.cq.FC6)
	f.P7.value = int(f.rendered.P7)
	f.P7.quality = int(f.rendered.cq.P7)
	f.T8.value = int(f.rendered.T8)
	f.T8.quality = int(f.rendered.cq.T8)
	f.F7.value = int(f.rendered.F7)
	f.F7.quality = int(f.rendered.cq.F7)
	f.F8.value = int(f.rendered.F8)
	f.F8.quality = int(f.rendered.cq.F8)
	f.T7.value = int(f.rendered.T7)
	f.T7.quality = int(f.rendered.cq.T7)
	f.P8.value = int(f.rendered.P8)
	f.P8.quality = int(f.rendered.cq.P8)
	f.AF4.value = int(f.rendered.AF4)
	f.AF4.quality = int(f.rendered.cq.AF4)
	f.F4.value = int(f.rendered.F4)
	f.F4.quality = int(f.rendered.cq.F4)
	f.AF3.value = int(f.rendered.AF3)
	f.AF3.quality = int(f.rendered.cq.AF3)
	f.O2.value = int(f.rendered.O2)
	f.O2.quality = int(f.rendered.cq.O2)
	f.O1.value = int(f.rendered.O1)
	f.O1.quality = int(f.rendered.cq.O1)
	f.FC5.value = int(f.rendered.FC5)
	f.FC5.quality = int(f.rendered.cq.FC5)
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
	f.parseSensors()
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

// Count returns the number of EPOC devices connected.
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

// Battery returns the current battery level of the device. May be 0 if
// the battery level has not yet been read.
func (f *EmokitFrame) Battery() uint {
	return uint(f.rendered.battery)
}

// Counter returns the counter value of the frame (0-127). 
func (f *EmokitFrame) Counter() uint {
	return uint(f.rendered.counter)
}
