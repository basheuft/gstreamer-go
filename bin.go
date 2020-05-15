package gstreamer

/*
#cgo pkg-config: gstreamer-1.0 gstreamer-base-1.0 gstreamer-app-1.0 gstreamer-plugins-base-1.0 gstreamer-video-1.0 gstreamer-audio-1.0 gstreamer-plugins-bad-1.0
#include "gstreamer.h"
*/
import "C"
import (
	"errors"
	"unsafe"
)

type Bin struct {
	bin *C.GstElement
}

func NewBin(name string) (*Bin, error) {
	nameUnsafe := C.CString(name)
	defer C.free(unsafe.Pointer(nameUnsafe))

	cBin := C.gst_bin_new(nameUnsafe)
	if cBin == nil {
		return nil, errors.New("create bin error")
	}

	bin := &Bin{
		bin: cBin,
	}

	return bin, nil
}

func (bin *Bin) AddElement(element *Element) {
	C.gstreamer_bin_add_element(bin.bin, element.element)
}

func (bin *Bin) FindByName(name string) (element *Element, found bool) {
	nameUnsafe := C.CString(name)
	defer C.free(unsafe.Pointer(nameUnsafe))

	el := C.gstreamer_bin_get_by_name(bin.bin, nameUnsafe)
	if el == nil {
		return nil, false
	}

	return &Element{
		element: el,
	}, true
}
