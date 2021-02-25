// +build darwin

package avfoundation

// #import <AVFoundation/AVFoundation.h>
import "C"
import "github.com/zyxar/mediastream/lib/format"

var pixelFormats = map[format.PixelFormat]C.FourCharCode{
	format.I420: C.kCVPixelFormatType_420YpCbCr8Planar,
	format.NV12: C.kCVPixelFormatType_420YpCbCr8BiPlanarVideoRange, // or C.kCVPixelFormatType_420YpCbCr8BiPlanarFullRange,
	format.UYVY: C.kCVPixelFormatType_422YpCbCr8,
	format.YUY2: C.kCVPixelFormatType_422YpCbCr8_yuvs,
	format.I444: C.kCVPixelFormatType_444YpCbCr8,
	format.RAW:  C.kCVPixelFormatType_24RGB,
	format.BGRA: C.kCVPixelFormatType_32ARGB,
	format.ARGB: C.kCVPixelFormatType_32BGRA,
}

func pixelFormatToFourCharCode(pf format.PixelFormat) (c C.FourCharCode, ok bool) {
	c, ok = pixelFormats[pf]
	return
}

func fourCharCodeToPixelFormat(c C.FourCharCode) (pf format.PixelFormat, ok bool) {
	for f, cc := range pixelFormats {
		if cc == c {
			return f, true
		}
	}
	return
}
