# media stream utility

## Prerequisites 

- `openh264` (for h264 encoding)
- `libvpx` (for vp8/vp9 encoding)
- `gstreamer` (and plugins)
    - `brew install gstreamer gst-plugins-good gst-plugins-base gst-plugins-ugly gst-plugins-bad gst-libav`

## RTP - H264

Launch an RTP server with h264, on port `5000`:

```shell
gst-launch-1.0 udpsrc port=5000 caps=application/x-rtp,encode-name=H264 \
       ! rtph264depay ! avdec_h264 ! videoconvert ! autovideosink
```

Run `mediastream` to stream video:
```shell
./mediastream -out rtp://127.0.0.1:5000 -codec h264
```

## RTP - VP8

- launch an RTP server with vp8, on port `5000`:

```shell
gst-launch-1.0 udpsrc port=5000 caps=application/x-rtp,encode-name=VP8 \
          ! rtpvp8depay ! avdec_vp8 ! videoconvert ! autovideosink
```

Run `mediastream` to stream video:
```shell
./mediastream -out rtp://127.0.0.1:5000 -codec vp8
```