package video

// credit: `github.com/pion/mediadevices/pkg/frame'

import (
	"errors"
	"fmt"

	"github.com/zyxar/mediastream/lib/format"
)

var decoders = map[PixelFormat]decoder{
	format.I420: decodeI420,
	format.I444: decodeI444,
	format.NV21: decodeNV21,
	format.NV12: decodeNV12,
	format.YUY2: decodeYUY2,
	format.UYVY: decodeUYVY,
	format.ARGB: decodeARGB,
	format.BGRA: decodeBGRA,
}

func Decode(f PixelFormat, buf []byte, width, height int) (Frame, error) {
	if decode, ok := decoders[f]; ok {
		return decode(buf, width, height)
	}
	return nil, fmt.Errorf("no decoder found for pixel format %q", f)
}

type decoder func(buf []byte, width, height int) (Frame, error)

var ErrInsufficientFrameBuffer = errors.New("insufficient frame buffer")

func DecodeToYUV420(f PixelFormat, buf []byte, width, height int) (Frame, error) {
	frame, err := Decode(f, buf, width, height)
	if err != nil {
		return nil, err
	}
	return Convert(frame)
}
