#pragma once

#import <AVFoundation/AVFoundation.h>
#include <pthread.h>

typedef struct
{
    FourCharCode pixelFormat;
    int width, height;
    double frameRate;
} VideoProperty;


typedef struct
{
    pthread_mutex_t          lock;
    VideoProperty            property;
    AVCaptureSession         *session;
    AVCaptureVideoDataOutput *output;
    CMSampleBufferRef        buffer;
    void                     *delegate;
} CaptureSession;
