package image

import (
	"errors"
	"image"
)

var ErrUnsupportedPixelFormat = errors.New("unsupported pixel format")

func Convert(src image.Image) (*image.YCbCr, error) {
	switch img := src.(type) {
	case *image.RGBA: // not yet
	case *image.YCbCr:
		switch img.SubsampleRatio {
		case image.YCbCrSubsampleRatio444:
			h := img.Rect.Dy()
			addrSrc0 := 0
			addrSrc1 := img.CStride
			addrDst := 0
			for i := 0; i < h/2; i++ {
				for j := 0; j < img.CStride/2; j++ {
					cb := uint16(img.Cb[addrSrc0]) + uint16(img.Cb[addrSrc1]) +
						uint16(img.Cb[addrSrc0+1]) + uint16(img.Cb[addrSrc1+1])
					cr := uint16(img.Cr[addrSrc0]) + uint16(img.Cr[addrSrc1]) +
						uint16(img.Cr[addrSrc0+1]) + uint16(img.Cr[addrSrc1+1])
					img.Cb[addrDst] = uint8(cb / 4)
					img.Cr[addrDst] = uint8(cr / 4)
					addrSrc0 += 2
					addrSrc1 += 2
					addrDst++
				}
				addrSrc0 += img.CStride
				addrSrc1 += img.CStride
			}
			img.CStride = img.CStride / 2
			cLen := img.CStride * (h / 2)
			img.Cb = img.Cb[:cLen]
			img.Cr = img.Cr[:cLen]
			img.SubsampleRatio = image.YCbCrSubsampleRatio420
			return img, nil
		case image.YCbCrSubsampleRatio422:
			h := img.Rect.Dy()
			addrSrc := 0
			addrDst := 0
			for i := 0; i < h/2; i++ {
				for j := 0; j < img.CStride; j++ {
					cb := uint16(img.Cb[addrSrc]) + uint16(img.Cb[addrSrc+img.CStride])
					cr := uint16(img.Cr[addrSrc]) + uint16(img.Cr[addrSrc+img.CStride])
					img.Cb[addrDst] = uint8(cb / 2)
					img.Cr[addrDst] = uint8(cr / 2)
					addrDst++
					addrSrc++
				}
				addrSrc += img.CStride
			}
			cLen := img.CStride * (h / 2)
			img.Cb = img.Cb[:cLen]
			img.Cr = img.Cr[:cLen]
			img.SubsampleRatio = image.YCbCrSubsampleRatio420
			return img, nil
		case image.YCbCrSubsampleRatio420:
			return img, nil
		}
	}

	return nil, ErrUnsupportedPixelFormat
}
