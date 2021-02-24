package format

type PixelFormat string

const ( // ref: https://chromium.googlesource.com/libyuv/libyuv/+show/master/docs/formats.md
	// Primary YUV formats:
	I420 PixelFormat = "I420"
	I422 PixelFormat = "I422"
	I444 PixelFormat = "I444"
	NV21 PixelFormat = "NV21"
	NV12 PixelFormat = "NV12"
	YUY2 PixelFormat = "YUY2"
	UYVY PixelFormat = "UYVY"

	// Primary RGB formats:
	ARGB PixelFormat = "ARGB"
	BGRA PixelFormat = "BGRA"
	RAW  PixelFormat = "RAW"
	RGBA PixelFormat = "RGBA"

	// Primary Compressed YUV format.
	MJPG PixelFormat = "MJPG"

	// Auxiliary aliases.
	IYUV = I420 // Alias for I420.
	YU16 = I422 // Alias for I422.
	YU24 = I444 // Alias for I444.
	YUYV = YUY2 // Alias for YUY2.
	YUVS = YUY2 // Alias for YUY2 on Mac.
	JPEG = MJPG // Alias for MJPG.
	RGB3 = RAW  // Alias for RAW.
	CM32 = BGRA // Alias for BGRA kCVPixelFormatType_32ARGB
	CM24 = RAW  // Alias for RAW kCVPixelFormatType_24RGB
)
