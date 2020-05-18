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

type Element struct {
	element *C.GstElement
	out     chan []byte
	stop    bool
	id      int
}

func NewElement(elType, name string) (element *Element, err error) {
	elTypeUnsafe := C.CString(elType)
	nameUnsafe := C.CString(name)
	defer C.free(unsafe.Pointer(elTypeUnsafe))
	defer C.free(unsafe.Pointer(nameUnsafe))

	cElement := C.gst_element_factory_make(elTypeUnsafe, nameUnsafe)
	if cElement == nil {
		return nil, errors.New("create element error")
	}

	element = &Element{
		element: cElement,
	}

	return element, nil
}

func (e *Element) SetCap(cap string) {
	capStr := C.CString(cap)
	defer C.free(unsafe.Pointer(capStr))
	C.gstreamer_set_caps(e.element, capStr)
}

func (e *Element) SetPropertyFloat(key string, val float32) {
	cKey := C.CString(key)
	cVal := C.float(val)
	defer C.free(unsafe.Pointer(cKey))

	C.gstreamer_set_property_float(e.element, cKey, cVal)
}

func (e *Element) Push(buffer []byte) {

	b := C.CBytes(buffer)
	defer C.free(unsafe.Pointer(b))
	C.gstreamer_element_push_buffer(e.element, b, C.int(len(buffer)))
}

func (e *Element) Poll() <-chan []byte {
	if e.out == nil {
		e.out = make(chan []byte, 10)
		C.gstreamer_element_pull_buffer(e.element, C.int(e.id))
	}
	return e.out
}

func (e *Element) Stop() {
	gstreamerLock.Lock()
	delete(elements, e.id)
	gstreamerLock.Unlock()
	if e.stop {
		return
	}
	if e.out != nil {
		e.stop = true
		close(e.out)
	}

}

func (e *Element) QueryDuration() int64 {
	duration := C.gstreamer_element_query_duration(e.element)
	return int64(duration)
}

func (e *Element) QueryPosition() int64 {
	position := C.gstreamer_element_query_position(e.element)
	return int64(position)
}
