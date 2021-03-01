package openh264

// #cgo pkg-config: openh264
/*
#include <stdint.h>
#include <stddef.h>
#include <wels/codec_api.h>

int newEncoder(ISVCEncoder **enc, int width, int height, int bitrate, float frameRate);
void closeEncoder(ISVCEncoder* enc);
int encode(ISVCEncoder *enc, uint8_t *dst, size_t *size, uint8_t *srcY, uint8_t *srcCb, uint8_t *srcCr, int width, int height);
int forceIntraFrame(ISVCEncoder *enc);
*/
import "C"
import (
	"image"
	"syscall"
)

type encoder struct{ enc *C.ISVCEncoder }

func NewEncoder(width int, height int, bitrate int, frameRate float64) (*encoder, error) {
	var enc *C.ISVCEncoder
	r := C.newEncoder(&enc, C.int(width), C.int(height), C.int(bitrate), C.float(frameRate))
	if r != 0 {
		return nil, syscall.EINVAL
	}
	return &encoder{enc: enc}, nil
}

func (e *encoder) Close() { C.closeEncoder(e.enc) }

func (e *encoder) encodeYUVFrame(dst []byte, i *image.YCbCr) (int, error) {
	var size C.size_t
	bounds := i.Bounds()
	r := C.encode(e.enc, (*C.uchar)(&dst[0]), &size,
		(*C.uchar)(&i.Y[0]),
		(*C.uchar)(&i.Cb[0]),
		(*C.uchar)(&i.Cr[0]),
		C.int(bounds.Max.X-bounds.Min.X),
		C.int(bounds.Max.Y-bounds.Min.Y),
	)
	if r != 0 {
		return 0, syscall.EINVAL
	}
	return int(size), nil
}

func (e *encoder) EncodeFrame(dst []byte, i image.Image) (int, error) {
	switch j := i.(type) {
	case *image.YCbCr:
		return e.encodeYUVFrame(dst, j)
	}
	panic("not implemented")
}

func (e *encoder) ForceIntraFrame() error {
	if C.forceIntraFrame(e.enc) != 0 {
		return syscall.EINVAL
	}
	return nil
}
