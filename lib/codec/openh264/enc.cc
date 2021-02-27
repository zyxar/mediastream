#include <stdint.h>
#include <stddef.h>
#include <string.h>
#include <wels/codec_api.h>

extern "C" {
int newEncoder(ISVCEncoder **enc, int width, int height, int bitrate, float frameRate);
void closeEncoder(ISVCEncoder* enc);
int encode(ISVCEncoder *enc, uint8_t *dst, size_t *size, uint8_t *srcY, uint8_t *srcCb, uint8_t *srcCr, int width, int height);
}

/* ref:
   https://ffmpeg.org/doxygen/2.6/libopenh264enc_8c_source.html
   https://github.com/cisco/openh264/wiki/UsageExampleForEncoder#encoder-usage-example-1
*/

int newEncoder(ISVCEncoder **enc, int width, int height, int bitrate, float frameRate)
{
    int ret;
    SEncParamExt param;
    ISVCEncoder* pEnc;
    int videoFormat = videoFormatI420;

    ret = WelsCreateSVCEncoder(&pEnc);
    if (ret != 0) {
        return ret;
    }
    ret = pEnc->GetDefaultParams(&param);
    if (ret != 0) {
        goto fail;
    }

    param.iUsageType = CAMERA_VIDEO_REAL_TIME;
    param.fMaxFrameRate = frameRate;
    param.iPicWidth = width;
    param.iPicHeight = height;
    param.iTargetBitrate = bitrate;
    param.iMaxBitrate = bitrate;
    param.iRCMode = RC_BITRATE_MODE; // RC_QUALITY_MODE;
    param.iTemporalLayerNum          = 1;
    param.iSpatialLayerNum           = 1;
    param.bEnableDenoise             = 0;
    param.bEnableBackgroundDetection = 1;
    param.bEnableAdaptiveQuant       = 1;
    param.bEnableFrameSkip           = 1;
    param.bEnableLongTermReference   = 0;
    param.iLtrMarkPeriod             = 30;
    param.bPrefixNalAddingCtrl       = 0;
    param.iEntropyCodingModeFlag     = 0;
    param.iMultipleThreadIdc         = 0;
    param.sSpatialLayers[0].iVideoWidth         = param.iPicWidth;
    param.sSpatialLayers[0].iVideoHeight        = param.iPicHeight;
    param.sSpatialLayers[0].fFrameRate          = param.fMaxFrameRate;
    param.sSpatialLayers[0].iSpatialBitrate     = param.iTargetBitrate;
    param.sSpatialLayers[0].iMaxSpatialBitrate  = param.iMaxBitrate;
    param.sSpatialLayers[0].sSliceArgument.uiSliceNum            = 1;
    param.sSpatialLayers[0].sSliceArgument.uiSliceMode           = SM_SIZELIMITED_SLICE;
    param.sSpatialLayers[0].sSliceArgument.uiSliceSizeConstraint = 12800;
    ret = pEnc->InitializeExt(&param);
    if (ret != cmResultSuccess) {
        goto fail;
    }
    pEnc->SetOption(ENCODER_OPTION_DATAFORMAT, &videoFormat);
    *enc = pEnc;
    return 0;

fail:
    WelsDestroySVCEncoder(pEnc);
    return ret;
}

void closeEncoder(ISVCEncoder* enc)
{
    if (enc) {
        enc->Uninitialize();
        WelsDestroySVCEncoder(enc);
    }
}

int encode(ISVCEncoder *enc, uint8_t *dst, size_t *size, uint8_t *srcY, uint8_t *srcCb, uint8_t *srcCr, int width, int height)
{
    int layer_size[MAX_LAYER_NUM_OF_FRAME] = { 0 };
    SFrameBSInfo fbi = { 0 };
    SSourcePicture sp = { 0 };
    sp.iColorFormat = videoFormatI420;
    sp.iPicWidth  = width;
    sp.iPicHeight = height;
    sp.iStride[0] = sp.iPicWidth;
    sp.iStride[1] = sp.iPicWidth / 2;
    sp.iStride[2] = sp.iPicWidth / 2;
    sp.pData[0] = srcY;
    sp.pData[1] = srcCb;
    sp.pData[2] = srcCr;
    int ret = enc->EncodeFrame(&sp, &fbi);
    if (ret != cmResultSuccess) {
        return ret;
    }
    *size = 0;
    if (fbi.eFrameType == videoFrameTypeSkip) {
        return 0;
    }
    for (int layer = 0; layer < fbi.iLayerNum; layer++) {
        for (int i = 0; i < fbi.sLayerInfo[layer].iNalCount; i++) {
            layer_size[layer] += fbi.sLayerInfo[layer].pNalLengthInByte[i];
        }
        memcpy(dst, fbi.sLayerInfo[layer].pBsBuf, layer_size[layer]);
        *size += layer_size[layer];
        dst += layer_size[layer];
    }
    return 0;
}