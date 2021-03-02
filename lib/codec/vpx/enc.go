package vpx

// #cgo pkg-config: vpx
/*
#include <stdlib.h>
#include <string.h>
#include <vpx/vpx_encoder.h>
#include <vpx/vpx_image.h>
#include <vpx/vp8cx.h>

int copyFrame(vpx_codec_ctx_t *ctx, uint8_t *dst)
{
    const vpx_codec_cx_pkt_t *pkt = NULL;
	vpx_codec_iter_t iter = NULL;
    int size = 0;

	while ((pkt = vpx_codec_get_cx_data(ctx, &iter))) {
		switch (pkt->kind) {
		case VPX_CODEC_CX_FRAME_PKT:
			memcpy(dst, pkt->data.frame.buf, pkt->data.frame.sz);
			dst += pkt->data.frame.sz;
            size += pkt->data.frame.sz;
			break;
        case VPX_CODEC_STATS_PKT:
			break;
        case VPX_CODEC_PSNR_PKT:
			break;
        case VPX_CODEC_CUSTOM_PKT:
			break;
		case VPX_CODEC_FPMB_STATS_PKT:
			break;
		}
	}

    return size;
}

vpx_codec_err_t initEncoder(vpx_codec_ctx_t **ctx, vpx_image_t **img, vpx_codec_enc_cfg_t *cfg, vpx_codec_iface_t *codec,
	unsigned int width, unsigned int height, unsigned int bitrate, unsigned int keyFrameInterval, int frameRate)
{
	vpx_codec_err_t e = vpx_codec_enc_config_default(codec, cfg, 0);
	if (e != VPX_CODEC_OK) {
		return e;
	}
	cfg->g_w = width;
	cfg->g_h = height;
	cfg->g_timebase.num = 1;
	cfg->g_timebase.den = frameRate;
	cfg->g_error_resilient = 1;
	cfg->g_pass = VPX_RC_ONE_PASS;
	cfg->rc_target_bitrate = bitrate;
	cfg->rc_resize_allowed = 0;
	cfg->kf_max_dist = keyFrameInterval;

	vpx_image_t i = {0};
	if (!vpx_img_alloc(&i, VPX_IMG_FMT_I420, width, height, 1)) {
		return VPX_CODEC_MEM_ERROR;
	}
	*img = calloc(1, sizeof(vpx_image_t));
	if (!*img) {
		return VPX_CODEC_MEM_ERROR;
	}
	**img = i;
	vpx_img_free(&i);
	*ctx = calloc(1, sizeof(vpx_codec_ctx_t));
	if (!*ctx) {
		free(*img);
		return VPX_CODEC_MEM_ERROR;
	}
	return vpx_codec_enc_init_ver(*ctx, codec, cfg, 0, VPX_ENCODER_ABI_VERSION);
}
*/
import "C"
import (
	"image"
	"sync/atomic"
	"unsafe"
)

type vpxError C.vpx_codec_err_t

func (v vpxError) Error() string {
	switch C.vpx_codec_err_t(v) {
	case C.VPX_CODEC_ERROR:
		return "CODEC_ERROR"
	case C.VPX_CODEC_MEM_ERROR:
		return "CODEC_MEM_ERROR"
	case C.VPX_CODEC_ABI_MISMATCH:
		return "CODEC_ABI_MISMATCH"
	case C.VPX_CODEC_INCAPABLE:
		return "CODEC_INCAPABLE"
	case C.VPX_CODEC_UNSUP_BITSTREAM:
		return "CODEC_UNSUP_BITSTREAM"
	case C.VPX_CODEC_UNSUP_FEATURE:
		return "CODEC_UNSUP_FEATURE"
	case C.VPX_CODEC_CORRUPT_FRAME:
		return "CODEC_CORRUPT_FRAME"
	case C.VPX_CODEC_INVALID_PARAM:
		return "CODEC_INVALID_PARAM"
	case C.VPX_CODEC_LIST_END:
		return "CODEC_LIST_END"
	}
	return "CODEC_UNKNOWN_ERR"
}

func codecError(c C.vpx_codec_err_t) error {
	switch c {
	case C.VPX_CODEC_OK:
		return nil
	default:
		return vpxError(c)
	}
}

type encoder struct {
	ctx              *C.vpx_codec_ctx_t
	img              *C.vpx_image_t
	cfg              C.vpx_codec_enc_cfg_t
	frameFlags       uint32 // vpx_enc_frame_flags_t
	frameCount       int64
	keyFrameInterval int
}

func NewVP8Encoder(width int, height int, bitrate int, keyFrameInterval int, frameRate float64) (*encoder, error) {
	var enc encoder
	err := C.initEncoder(&enc.ctx, &enc.img, &enc.cfg, C.vpx_codec_vp8_cx(),
		C.uint(width), C.uint(height), C.uint(bitrate/1000), C.uint(keyFrameInterval), C.int(frameRate))
	if err != C.VPX_CODEC_OK {
		return nil, codecError(err)
	}
	enc.keyFrameInterval = keyFrameInterval
	return &enc, nil
}

func NewVP9Encoder(width int, height int, bitrate int, keyFrameInterval int, frameRate float64) (*encoder, error) {
	var enc encoder
	err := C.initEncoder(&enc.ctx, &enc.img, &enc.cfg, C.vpx_codec_vp9_cx(),
		C.uint(width), C.uint(height), C.uint(bitrate/1000), C.uint(keyFrameInterval), C.int(frameRate))
	if err != C.VPX_CODEC_OK {
		return nil, codecError(err)
	}
	enc.keyFrameInterval = keyFrameInterval
	return &enc, nil
}

func (e *encoder) Close() error {
	C.free(unsafe.Pointer(e.img))
	err := codecError(C.vpx_codec_destroy(e.ctx))
	C.free(unsafe.Pointer(e.ctx))
	return err
}

func (e *encoder) ForceIntraFrame() error {
	for {
		oldVal := atomic.LoadUint32(&e.frameFlags)
		newVal := oldVal | uint32(C.VPX_EFLAG_FORCE_KF)
		if newVal == oldVal || atomic.CompareAndSwapUint32(&e.frameFlags, oldVal, newVal) {
			return nil
		}
	}
	return nil
}

func (e *encoder) EncodeFrame(dst []byte, i image.Image) (int, error) {
	switch j := i.(type) {
	case *image.YCbCr:
		flag := atomic.SwapUint32(&e.frameFlags, 0)
		return e.encodeYUVFrame(dst, flag, j)
	}
	panic("not implemented")
}

func (e *encoder) encodeYUVFrame(dst []byte, flag uint32, i *image.YCbCr) (int, error) {
	e.img.stride[0] = C.int(i.YStride)
	e.img.stride[1] = C.int(i.CStride)
	e.img.stride[2] = C.int(i.CStride)
	e.img.planes[0] = (*C.uchar)(&i.Y[0])
	e.img.planes[1] = (*C.uchar)(&i.Cb[0])
	e.img.planes[2] = (*C.uchar)(&i.Cr[0])
	if e.keyFrameInterval > 0 && e.frameCount%int64(e.keyFrameInterval) == 0 {
		flag |= C.VPX_EFLAG_FORCE_KF
	}
	// FIXME: on resolution change?
	err := C.vpx_codec_encode(e.ctx, e.img, C.vpx_codec_pts_t(e.frameCount), 1, C.vpx_enc_frame_flags_t(flag), C.VPX_DL_REALTIME)
	if err != C.VPX_CODEC_OK {
		return 0, codecError(err)
	}
	e.frameCount++
	size := C.copyFrame(e.ctx, (*C.uchar)(&dst[0]))
	return int(size), nil
}
