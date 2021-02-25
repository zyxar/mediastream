// +build darwin

package avfoundation

// #cgo CFLAGS: -x objective-c
// #cgo LDFLAGS: -framework AVFoundation -framework Foundation -framework CoreMedia -framework CoreVideo
/*
#include "avfoundation.h"

extern void closeSession(CaptureSession *s);
extern int initSession(CaptureSession *s);
extern int readVideoFrame(CaptureSession* s, uint8_t *buf, size_t size);
extern size_t getVideoBufferSize(CaptureSession *s);
*/
import "C"
import (
	"syscall"

	"github.com/zyxar/mediastream/lib/format"
)

type Property struct {
	format.PixelFormat
	Width, Height int
	FrameRate     float64
}

type Session struct {
	s       C.CaptureSession
	p       Property
	bufSize int
}

func NewSession(p Property) (*Session, error) {
	var s Session
	s.s.property.pixelFormat = pixelFormats[p.PixelFormat]
	s.s.property.width = C.int(p.Width)
	s.s.property.height = C.int(p.Height)
	s.s.property.frameRate = C.double(p.FrameRate)
	ret := C.initSession(&s.s)
	if ret != 0 {
		return nil, syscall.Errno(ret)
	}
	return s.init(), nil
}

func (s *Session) init() *Session {
	size := C.getVideoBufferSize(&s.s)
	s.bufSize = int(size)
	s.p.Width = int(s.s.property.width)
	s.p.Height = int(s.s.property.height)
	s.p.FrameRate = float64(s.s.property.frameRate)
	s.p.PixelFormat, _ = fourCharCodeToPixelFormat(s.s.property.pixelFormat)
	return s
}

func (s *Session) BufferSize() int    { return s.bufSize }
func (s *Session) Property() Property { return s.p }
func (s *Session) Close()             { C.closeSession(&s.s) }
func (s *Session) ReadVideoFrame(buf []byte) (i int, err error) {
	ret := C.readVideoFrame(&s.s, (*C.uchar)(&buf[0]), C.size_t(len(buf)))
	if ret < 0 {
		return 0, syscall.EAGAIN
	}
	return int(ret), nil
}
