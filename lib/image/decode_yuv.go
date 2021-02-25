package image

import "image"

func decodeI420(frame []byte, width, height int) (image.Image, error) {
	yi := width * height
	cbi := yi + yi/4
	cri := cbi + yi/4
	if cri > len(frame) {
		return nil, ErrInsufficientFrameBuffer
	}
	return &image.YCbCr{
		Y:              frame[:yi],
		YStride:        width,
		Cb:             frame[yi:cbi],
		Cr:             frame[cbi:cri],
		CStride:        width / 2,
		SubsampleRatio: image.YCbCrSubsampleRatio420,
		Rect:           image.Rect(0, 0, width, height),
	}, nil
}

func decodeYV12(frame []byte, width, height int) (image.Image, error) {
	img, err := decodeI420(frame, width, height)
	if err != nil {
		return img, err
	}
	yuv := img.(*image.YCbCr)
	yuv.Cb, yuv.Cr = yuv.Cr, yuv.Cb
	return yuv, err
}

func decodeI444(frame []byte, width, height int) (image.Image, error) {
	yi := width * height
	cbi := yi * 2
	cri := yi * 3
	if cri > len(frame) {
		return nil, ErrInsufficientFrameBuffer
	}
	return &image.YCbCr{
		Y:              frame[:yi],
		YStride:        width,
		Cb:             frame[yi:cbi],
		Cr:             frame[cbi:cri],
		CStride:        width,
		SubsampleRatio: image.YCbCrSubsampleRatio444,
		Rect:           image.Rect(0, 0, width, height),
	}, nil
}

func decodeYV24(frame []byte, width, height int) (image.Image, error) {
	yi := width * height
	cbi := yi * 2
	cri := yi * 3
	if cri > len(frame) {
		return nil, ErrInsufficientFrameBuffer
	}
	return &image.YCbCr{
		Y:              frame[:yi],
		YStride:        width,
		Cr:             frame[yi:cbi],
		Cb:             frame[cbi:cri],
		CStride:        width,
		SubsampleRatio: image.YCbCrSubsampleRatio444,
		Rect:           image.Rect(0, 0, width, height),
	}, nil
}

func decodeNV21(frame []byte, width, height int) (image.Image, error) {
	yi := width * height
	ci := yi + yi/2
	if ci > len(frame) {
		return nil, ErrInsufficientFrameBuffer
	}
	cb := make([]byte, 0, yi/2)
	cr := make([]byte, 0, yi/2)
	for i := yi; i < ci; i += 2 {
		cr = append(cr, frame[i])
		cb = append(cb, frame[i+1])
	}
	return &image.YCbCr{
		Y:              frame[:yi],
		YStride:        width,
		Cb:             cb,
		Cr:             cr,
		CStride:        width / 2,
		SubsampleRatio: image.YCbCrSubsampleRatio420,
		Rect:           image.Rect(0, 0, width, height),
	}, nil
}

func decodeNV12(frame []byte, width, height int) (image.Image, error) {
	yi := width * height
	ci := yi + yi/2
	if ci > len(frame) {
		return nil, ErrInsufficientFrameBuffer
	}
	cb := make([]byte, 0, yi/2)
	cr := make([]byte, 0, yi/2)
	for i := yi; i < ci; i += 2 {
		cb = append(cb, frame[i])
		cr = append(cr, frame[i+1])
	}
	return &image.YCbCr{
		Y:              frame[:yi],
		YStride:        width,
		Cb:             cb,
		Cr:             cr,
		CStride:        width / 2,
		SubsampleRatio: image.YCbCrSubsampleRatio420,
		Rect:           image.Rect(0, 0, width, height),
	}, nil
}

func decodeYUY2(frame []byte, width, height int) (image.Image, error) {
	yi := width * height
	ci := yi / 2
	fi := yi * 2
	if len(frame) < fi {
		return nil, ErrInsufficientFrameBuffer
	}
	y := make([]byte, yi)
	cb := make([]byte, ci)
	cr := make([]byte, ci)
	fillYUY2(y, cb, cr, frame, width, height)
	return &image.YCbCr{
		Y:              y,
		YStride:        width,
		Cb:             cb,
		Cr:             cr,
		CStride:        width / 2,
		SubsampleRatio: image.YCbCrSubsampleRatio422,
		Rect:           image.Rect(0, 0, width, height),
	}, nil
}

func decodeUYVY(frame []byte, width, height int) (image.Image, error) {
	yi := width * height
	ci := yi / 2
	fi := yi * 2
	if len(frame) < fi {
		return nil, ErrInsufficientFrameBuffer
	}
	y := make([]byte, yi)
	cb := make([]byte, ci)
	cr := make([]byte, ci)
	fillUYVY(y, cb, cr, frame, width, height)
	return &image.YCbCr{
		Y:              y,
		YStride:        width,
		Cb:             cb,
		Cr:             cr,
		CStride:        width / 2,
		SubsampleRatio: image.YCbCrSubsampleRatio422,
		Rect:           image.Rect(0, 0, width, height),
	}, nil
}
