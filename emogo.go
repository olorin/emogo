package emogo

import (
	"errors"
)

// #include <emokit/emokit.h>
// #include <stdint.h>
// #cgo LDFLAGS: -lemokit
import "C"

// These are defined in emokit.h and reproduced here as cgo isn't
// linking them in for some reason.
const (
	EMOKIT_VID int = 0x21a1
	EMOKIT_PID int = 0x0001
)

// EmokitContext represents a connection to an EPOC device. 
type EmokitContext struct {
	eeg *C.struct_emokit_device
}

func NewEmokitContext() (*EmokitContext, error) {
	e := new(EmokitContext)
	e.eeg = C.emokit_create()
	ret := C.emokit_open(e.eeg, C.int(EMOKIT_VID), C.int(EMOKIT_PID), 0)
	if ret != 0 {
		return nil, errors.New("Cannot access device.")
	}
	return e, nil
}
