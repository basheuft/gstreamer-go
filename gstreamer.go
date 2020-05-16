package gstreamer

/*
#cgo pkg-config: gstreamer-1.0 gstreamer-base-1.0 gstreamer-app-1.0 gstreamer-plugins-base-1.0 gstreamer-video-1.0 gstreamer-audio-1.0 gstreamer-plugins-bad-1.0
#include "gstreamer.h"
*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"
)

func init() {
	C.gstreamer_init()
}

type MessageType int

const (
	MESSAGE_UNKNOWN       MessageType = C.GST_MESSAGE_UNKNOWN
	MESSAGE_EOS           MessageType = C.GST_MESSAGE_EOS
	MESSAGE_ERROR         MessageType = C.GST_MESSAGE_ERROR
	MESSAGE_WARNING       MessageType = C.GST_MESSAGE_WARNING
	MESSAGE_INFO          MessageType = C.GST_MESSAGE_INFO
	MESSAGE_TAG           MessageType = C.GST_MESSAGE_TAG
	MESSAGE_BUFFERING     MessageType = C.GST_MESSAGE_BUFFERING
	MESSAGE_STATE_CHANGED MessageType = C.GST_MESSAGE_STATE_CHANGED
	MESSAGE_ANY           MessageType = C.GST_MESSAGE_ANY
)

type Message struct {
	GstMessage *C.GstMessage
}

func (v *Message) GetType() MessageType {
	c := C.toGstMessageType(unsafe.Pointer(v.native()))
	return MessageType(c)
}

func (v *Message) native() *C.GstMessage {
	if v == nil {
		return nil
	}
	return v.GstMessage
}

func (v *Message) GetTimestamp() uint64 {
	c := C.messageTimestamp(unsafe.Pointer(v.native()))
	return uint64(c)
}

func (v *Message) GetTypeName() string {
	c := C.messageTypeName(unsafe.Pointer(v.native()))
	return C.GoString(c)
}

func gbool(b bool) C.gboolean {
	if b {
		return C.gboolean(1)
	}
	return C.gboolean(0)
}
func gobool(b C.gboolean) bool {
	if b != 0 {
		return true
	}
	return false
}

var pipelines = make(map[int]*Pipeline)
var elements = make(map[int]*Element)
var gstreamerLock sync.Mutex
var gstreamerIdGenerate = 10000

//export goHandleSinkBuffer
func goHandleSinkBuffer(buffer unsafe.Pointer, bufferLen C.int, elementID C.int) {
	gstreamerLock.Lock()
	defer gstreamerLock.Unlock()
	if element, ok := elements[int(elementID)]; ok {
		if element.out != nil && !element.stop {
			element.out <- C.GoBytes(buffer, bufferLen)
		}
	} else {
		fmt.Printf("discarding buffer, no element with id %d", int(elementID))
	}
	C.free(buffer)
}

//export goHandleSinkEOS
func goHandleSinkEOS(elementID C.int) {
	gstreamerLock.Lock()
	defer gstreamerLock.Unlock()
	if element, ok := elements[int(elementID)]; ok {
		if element.out != nil && !element.stop {
			element.stop = true
			close(element.out)
		}
	}
}

//export goHandleBusMessage
func goHandleBusMessage(message *C.GstMessage, pipelineId C.int) {
	//log.Printf("MESSAGE: %v - %v", pipelineId, message)

	gstreamerLock.Lock()
	defer gstreamerLock.Unlock()
	msg := &Message{GstMessage: message}
	if pipeline, ok := pipelines[int(pipelineId)]; ok {
		if pipeline.messages != nil {
			pipeline.messages <- msg
		}
	} else {
		fmt.Printf("discarding message, no pipelie with id %d", int(pipelineId))
	}

}

// ScanPathForPlugins : Scans a given path for any gstreamer plugins and adds them to
// the gst_registry
func ScanPathForPlugins(directory string) {
	C.gst_registry_scan_path(C.gst_registry_get(), C.CString(directory))
}

func CheckPlugins(plugins []string) error {

	var plugin *C.GstPlugin
	var registry *C.GstRegistry

	registry = C.gst_registry_get()

	for _, pluginstr := range plugins {
		plugincstr := C.CString(pluginstr)
		plugin = C.gst_registry_find_plugin(registry, plugincstr)
		C.free(unsafe.Pointer(plugincstr))
		if plugin == nil {
			return fmt.Errorf("Required gstreamer plugin %s not found", pluginstr)
		}
	}

	return nil
}
