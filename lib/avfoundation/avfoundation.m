#include "avfoundation.h"

@interface CaptureDelegate : NSObject<AVCaptureVideoDataOutputSampleBufferDelegate>
{
    CaptureSession* cs;
}

- (void)  captureOutput:(AVCaptureOutput *)captureOutput
  didOutputSampleBuffer:(CMSampleBufferRef)frameBuffer
         fromConnection:(AVCaptureConnection *)connection;

@end

@implementation CaptureDelegate

- (id) init: (CaptureSession*) s
{
    self = [super init];
    self->cs = s;
    return self;
}

- (void)  captureOutput:(AVCaptureOutput *)captureOutput
  didOutputSampleBuffer:(CMSampleBufferRef)frameBuffer
         fromConnection:(AVCaptureConnection *)connection
{
    pthread_mutex_lock(&self->cs->lock);
    if (self->cs->buffer != nil) {
        CFRelease(self->cs->buffer);
    }
    self->cs->buffer = (CMSampleBufferRef)CFRetain(frameBuffer);
    pthread_mutex_unlock(&self->cs->lock);
}

@end

static int initDevice(AVCaptureDevice *device, VideoProperty* vp)
{
    NSObject *range = nil;
    NSObject *format = nil;
    NSObject *selectedRange = nil;
    NSObject *selectedFormat = nil;

    @try {
        for (format in [device valueForKey:@"formats"]) {
            CMFormatDescriptionRef formatDescription = (CMFormatDescriptionRef) [format performSelector:@selector(formatDescription)];
            CMVideoDimensions dimensions = CMVideoFormatDescriptionGetDimensions(formatDescription);
            if ((vp->width == 0 && vp->height == 0) || (dimensions.width == vp->width && dimensions.height == vp->height)) {
                selectedFormat = format;
                for (range in [format valueForKey:@"videoSupportedFrameRateRanges"]) {
                    double maxFrameRate;
                    [[range valueForKey:@"maxFrameRate"] getValue:&maxFrameRate];
                    if (fabs (vp->frameRate - maxFrameRate) < 0.01) {
                        selectedRange = range;
                        break;
                    }
                }
            }
        }

        if (!selectedFormat) {
            NSLog(@"Selected video size (%dx%d) is not supported by the device.", vp->width, vp->height);
            return EINVAL;
        }
        if (!selectedRange && vp->frameRate > 0) {
            NSLog(@"Selected framerate (%f) is not supported by the device.", vp->frameRate);
            return EINVAL;
        }
        if (vp->width == 0 && vp->height == 0) {
            CMVideoFormatDescriptionRef desc = (CMFormatDescriptionRef) [selectedFormat performSelector:@selector(formatDescription)];
            CMVideoDimensions dim = CMVideoFormatDescriptionGetDimensions(desc);
            vp->height = dim.height;
            vp->width = dim.width;
        }
        if ([device lockForConfiguration:nil] == YES) {
            if (selectedFormat) {
                [device setValue:selectedFormat forKey:@"activeFormat"];
            }
            if (selectedRange) {
                NSValue *minFrameDuration = [selectedRange valueForKey:@"minFrameDuration"];
                [device setValue:minFrameDuration forKey:@"activeVideoMinFrameDuration"];
                [device setValue:minFrameDuration forKey:@"activeVideoMaxFrameDuration"];
            }
        } else {
            NSLog(@"Could not lock device for configuration.");
            return EINVAL;
        }
    } @catch(NSException *e) {
        NSLog(@"Configuration of video device failed: %s", [e.reason UTF8String]);
    }

    return 0;
}

void closeSession(CaptureSession *s)
{
    [s->session stopRunning];
    [s->session release];
    [s->output release];
    [(CaptureDelegate*)(s->delegate) release];

    s->session = nil;
    s->output = nil;
    s->delegate = nil;

    pthread_mutex_destroy(&(s->lock));

    if (s->buffer) {
        CFRelease(s->buffer);
        s->buffer = nil;
    }
}


int initSession(CaptureSession *s)
{
    NSError *error = nil;
    AVCaptureInput* input = nil;
    NSNumber *pixelFormat = [NSNumber numberWithUnsignedInt:s->property.pixelFormat];
    NSDictionary *setting = nil;
    CaptureDelegate *delegate = nil;
    dispatch_queue_t queue;

    NSAutoreleasePool *pool = [[NSAutoreleasePool alloc] init];
    pthread_mutex_init(&(s->lock), nil);

    AVCaptureDevice *device = [AVCaptureDevice defaultDeviceWithMediaType:AVMediaTypeVideo];
    if (device == nil)
        goto fail;

    s->session = [[AVCaptureSession alloc] init];
    input = (AVCaptureInput*) [[[AVCaptureDeviceInput alloc] initWithDevice:device error:&error] autorelease];
    if (!input) {
        NSLog(@"Failed to create capture input device: %s", [[error localizedDescription] UTF8String]);
        goto fail;
    }
    if ([s->session canAddInput:input]) {
        [s->session addInput:input];
    } else {
        NSLog(@"Cannot add video input to capture session");
        goto fail;
    }
    s->output = [[AVCaptureVideoDataOutput alloc] init];
    if (!s->output) {
        NSLog(@"Failed to init video output");
        goto fail;
    }

    if (initDevice(device, &s->property)) {
        NSLog(@"Failed to init video device");
        goto fail;
    }

    if ([[s->output availableVideoCVPixelFormatTypes] indexOfObject:pixelFormat] == NSNotFound) {
        NSLog(@"Selected pixel format (%08x) is not supported by the input device.", [pixelFormat intValue]);
        pixelFormat = [[s->output availableVideoCVPixelFormatTypes] objectAtIndex:0]; // FIXME: choose supported one
        s->property.pixelFormat = [pixelFormat intValue];
        NSLog(@"Overriding selected pixel format to use %08x instead", s->property.pixelFormat);
    }
    setting = [NSDictionary dictionaryWithObject:pixelFormat
                                          forKey:(id)kCVPixelBufferPixelFormatTypeKey];
    [s->output setVideoSettings:setting];
    [s->output setAlwaysDiscardsLateVideoFrames:YES];

    delegate = [[CaptureDelegate alloc] init:s];
    queue = dispatch_queue_create("avf_queue", nil);
    [s->output setSampleBufferDelegate:delegate queue:queue];
    dispatch_release(queue);
    s->delegate = delegate;

    if ([s->session canAddOutput:s->output]) {
        [s->session addOutput:s->output];
    } else {
        NSLog(@"Cannot add video output to capture session");
        goto fail;
    }

    [s->session startRunning];
    [device unlockForConfiguration];

    [pool release];
    return 0;

fail:
    [pool release];
    closeSession(s);
    return EIO;
}


int readVideoFrame(CaptureSession* s, uint8_t *buf, size_t size)
{
    void *data = nil;
    int len = 0;
    int status = 0;
    do {
        CVImageBufferRef imageBuffer;
        pthread_mutex_lock(&s->lock);
        if (s->buffer != nil) {
            imageBuffer = CMSampleBufferGetImageBuffer(s->buffer);
            if (imageBuffer) {
                status = CVPixelBufferLockBaseAddress(imageBuffer, 0);
                if (status == kCVReturnSuccess) {
                    if (CVPixelBufferIsPlanar(imageBuffer)) {
                        size_t count = CVPixelBufferGetPlaneCount(imageBuffer);
                        int planeSize = 0;
                        for (int i = 0; i < count; i++) {
                            data = CVPixelBufferGetBaseAddressOfPlane(imageBuffer, i);
                            planeSize = CVPixelBufferGetBytesPerRowOfPlane(imageBuffer, i) *
                                        CVPixelBufferGetHeightOfPlane(imageBuffer, i);
                            memcpy(buf, data, planeSize);
                            buf += planeSize;
                            len += planeSize;
                        }
                    } else {
                        data = CVPixelBufferGetBaseAddress(imageBuffer);
                        len = CVPixelBufferGetBytesPerRow(imageBuffer) * CVPixelBufferGetHeight(imageBuffer);
                        // FIXME: ensure len < CVPixelBufferGetDataSize(imageBuffer)
                        memcpy(buf, data, len);
                    }
                    CVPixelBufferUnlockBaseAddress(imageBuffer, 0);
                }
            }
            CFRelease(s->buffer);
            s->buffer = nil;
        }
        pthread_mutex_unlock(&s->lock);
    } while (!data);

    if (status != kCVReturnSuccess) {
        return -1;
    }
    return len;
}

size_t getVideoBufferSize(CaptureSession *s)
{
    size_t size = 0;
    CMSampleBufferRef buffer = nil;
    do {
        CFRunLoopRunInMode(kCFRunLoopDefaultMode, 0.1, YES);
        pthread_mutex_lock(&s->lock);
        if (s->buffer != nil) {
            buffer = s->buffer;
            s->buffer = nil;
        }
        pthread_mutex_unlock(&s->lock);
    } while (!buffer);

    CVImageBufferRef imageBuffer = CMSampleBufferGetImageBuffer(buffer);
    if (imageBuffer) {
        size = CVPixelBufferGetDataSize(imageBuffer);
    }
    CFRelease(buffer);
    return size;
}


