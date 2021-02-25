package image

// credit: `github.com/pion/mediadevices/pkg/frame'

import (
	"errors"
	"fmt"
	"image"

	"github.com/zyxar/mediastream/lib/format"
)

type Image = image.Image

var decoders = map[format.PixelFormat]decoder{
	format.I420: decodeI420,
	format.I444: decodeI444,
	format.NV21: decodeNV21,
	format.NV12: decodeNV12,
	format.YUY2: decodeYUY2,
	format.UYVY: decodeUYVY,
	format.ARGB: decodeARGB,
	format.BGRA: decodeBGRA,
}

func Decode(f format.PixelFormat, buf []byte, width, height int) (Image, error) {
	if decode, ok := decoders[f]; ok {
		return decode(buf, width, height)
	}
	return nil, fmt.Errorf("no decoder found for pixel format %q", f)
}

type decoder func(buf []byte, width, height int) (Image, error)

var ErrInsufficientFrameBuffer = errors.New("insufficient frame buffer")
