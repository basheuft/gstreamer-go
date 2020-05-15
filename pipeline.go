package gstreamer

/*
#cgo pkg-config: gstreamer-1.0 gstreamer-base-1.0 gstreamer-app-1.0 gstreamer-plugins-base-1.0 gstreamer-video-1.0 gstreamer-audio-1.0 gstreamer-plugins-bad-1.0
#include "gstreamer.h"
*/
import "C"
import (
	"errors"
	"log"
	"unsafe"
)

type Pipeline struct {
	pipeline *C.GstPipeline
	pipelineEl *C.GstElement
	messages chan *Message
	id       int
}

func New(pipelineStr string) (*Pipeline, error) {
	pipelineStrUnsafe := C.CString(pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))
	cpipeline := C.gstreamer_create_pipeline(pipelineStrUnsafe)
	if cpipeline == nil {
		return nil, errors.New("create pipeline error")
	}

	pipeline := &Pipeline{
		pipeline: cpipeline,
	}

	gstreamerLock.Lock()
	defer gstreamerLock.Unlock()
	gstreamerIdGenerate += 1
	pipeline.id = gstreamerIdGenerate
	pipelines[pipeline.id] = pipeline
	return pipeline, nil
}

func (p *Pipeline) PullMessage() <-chan *Message {
	p.messages = make(chan *Message, 5)
	C.gstreamer_pipeline_but_watch(p.pipeline, C.int(p.id))
	return p.messages
}

func (p *Pipeline) Start() {
	log.Printf("Starting pipeline with ID: %v", p.id)
	C.gstreamer_pipeline_start(p.pipeline, C.int(p.id))
}

func (p *Pipeline) Pause() {
	C.gstreamer_pipeline_pause(p.pipeline)
}

func (p *Pipeline) Stop() {
	gstreamerLock.Lock()
	delete(pipelines, p.id)
	gstreamerLock.Unlock()
	if p.messages != nil {
		close(p.messages)
	}
	C.gstreamer_pipeline_stop(p.pipeline)
}

func (p *Pipeline) SendEOS() {
	C.gstreamer_pipeline_sendeos(p.pipeline)
}

func (p *Pipeline) SetAutoFlushBus(flush bool) {
	gflush := gbool(flush)
	C.gstreamer_pipeline_set_auto_flush_bus(p.pipeline, gflush)
}

func (p *Pipeline) GetAutoFlushBus() bool {
	gflush := C.gstreamer_pipeline_get_auto_flush_bus(p.pipeline)
	return gobool(gflush)
}

func (p *Pipeline) GetDelay() uint64 {

	delay := C.gstreamer_pipeline_get_delay(p.pipeline)
	return uint64(delay)
}

func (p *Pipeline) SetDelay(delay uint64) {
	C.gstreamer_pipeline_set_delay(p.pipeline, C.guint64(delay))
}

func (p *Pipeline) GetLatency() uint64 {

	latency := C.gstreamer_pipeline_get_latency(p.pipeline)
	return uint64(latency)
}

func (p *Pipeline) SetLatency(latency uint64) {
	C.gstreamer_pipeline_set_latency(p.pipeline, C.guint64(latency))
}

func (p *Pipeline) FindElement(name string) *Element {
	elementName := C.CString(name)
	defer C.free(unsafe.Pointer(elementName))
	gelement := C.gstreamer_pipeline_findelement(p.pipeline, elementName)
	if gelement == nil {
		return nil
	}
	element := &Element{
		element: gelement,
	}

	gstreamerLock.Lock()
	defer gstreamerLock.Unlock()
	gstreamerIdGenerate += 1
	element.id = gstreamerIdGenerate
	elements[element.id] = element

	return element
}

func (p *Pipeline) AddElement(name string, qName string) *Element {
	nameUnsafe := C.CString(name)
	qNameUnsafe := C.CString(qName)
	defer C.free(unsafe.Pointer(nameUnsafe))
	defer C.free(unsafe.Pointer(qNameUnsafe))

	el := C.gst_element_factory_make(nameUnsafe, qNameUnsafe)

	return &Element{
		element: el,
	}
}

func (p *Pipeline) AddBin(bin *Bin) {
	C.gstreamer_pipeline_bin_add(p.pipeline, bin.bin)
}
