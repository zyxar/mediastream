// +build darwin

package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"math"
	"math/rand"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/zyxar/mediastream/lib/avfoundation"
	"github.com/zyxar/mediastream/lib/codec/openh264"
	"github.com/zyxar/mediastream/lib/codec/vpx"
	"github.com/zyxar/mediastream/lib/format"
	"github.com/zyxar/mediastream/lib/video"

	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
)

var (
	selectedFormat    = flag.String("format", "NV12", "set pixel format")
	selectedFrameRate = flag.Float64("framerate", 30, "set frame rate")
	selectedOut       = flag.String("out", "", "set output file name")
	selectedCodec     = flag.String("codec", "h264", "set codec for output (h264/vp8/vp9)")
)

func main() {
	flag.Parse()

	var pixelFormat = format.PixelFormat(strings.ToUpper(*selectedFormat))
	s, err := avfoundation.NewSession(
		avfoundation.Property{PixelFormat: pixelFormat, Width: 640, Height: 480, FrameRate: *selectedFrameRate})
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	p := s.Property()

	var imageBuffer = make([]byte, s.BufferSize())
	var process = func(w io.Writer) error {
		if _, err := s.ReadVideoFrame(imageBuffer); err != nil {
			return err
		}
		_, err = w.Write(imageBuffer)
		return err
	}

	if *selectedOut != "" {
		var frameEncoder interface {
			EncodeFrame(dst []byte, i image.Image) (int, error)
		}
		var payloader rtp.Payloader
		var payloadType uint8
		switch strings.ToLower(*selectedCodec) {
		case "h264", "264":
			codec, err := openh264.NewEncoder(p.Width, p.Height, 500_000, p.FrameRate)
			if err != nil {
				log.Fatal(err)
			}
			defer codec.Close()
			frameEncoder = codec
			payloader = &codecs.H264Payloader{}
			payloadType = 125
		case "vp8":
			codec, err := vpx.NewVP8Encoder(p.Width, p.Height, 500_000, 60, p.FrameRate)
			if err != nil {
				log.Fatal(err)
			}
			defer codec.Close()
			frameEncoder = codec
			payloader = &codecs.VP8Payloader{}
			payloadType = 100
		case "vp9":
			codec, err := vpx.NewVP9Encoder(p.Width, p.Height, 500_000, 60, p.FrameRate)
			if err != nil {
				log.Fatal(err)
			}
			defer codec.Close()
			frameEncoder = codec
			payloader = &codecs.VP9Payloader{}
			payloadType = 101
		default:
			log.Fatalf("unsupported codec: %v", *selectedCodec)
		}

		var frameBuffer = make([]byte, s.BufferSize())
		enc := func(w io.Writer) writerFn {
			return func(buf []byte) (n int, err error) {
				img, err := video.DecodeToYUV420(p.PixelFormat, buf, p.Width, p.Height)
				if err != nil {
					return n, err
				}
				l, err := frameEncoder.EncodeFrame(frameBuffer, img)
				if l > 0 {
					return w.Write(frameBuffer[:l])
				}
				return n, nil
			}
		}

		var writer io.Writer
		uri, err := url.Parse(*selectedOut)
		if err != nil {
			log.Fatal(err)
		}
		switch uri.Scheme {
		case "rtp":
			conn, err := net.Dial("udp", uri.Host)
			if err != nil {
				log.Fatal(err)
			}
			defer conn.Close()
			writer = enc(newRTPWriter(conn, payloadType, payloader))
		default:
			file, err := os.Create(*selectedOut)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
			writer = enc(file)
		}

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGABRT, syscall.SIGTERM, syscall.SIGHUP)
		defer signal.Stop(sig)

		for {
			select {
			case <-sig:
				return
			default:
				if err = process(writer); err != nil {
					log.Println(err)
					return
				}
			}
		}

		os.Exit(0)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mimeWriter := multipart.NewWriter(w)
		contentType := fmt.Sprintf("multipart/x-mixed-replace;boundary=%s", mimeWriter.Boundary())
		w.Header().Add("Content-Type", contentType)
		partHeader := make(textproto.MIMEHeader)
		partHeader.Add("Content-Type", "image/jpeg")

		enc := func(w io.Writer) writerFn {
			return func(buf []byte) (n int, err error) {
				img, err := video.Decode(p.PixelFormat, buf, p.Width, p.Height)
				if err != nil {
					return n, err
				}
				return n, jpeg.Encode(w, img, nil) // FIXME: n?
			}
		}

		for {
			partWriter, err := mimeWriter.CreatePart(partHeader)
			if err != nil {
				log.Println(err)
				return
			}
			err = process(enc(partWriter))
			if err != nil {
				log.Println(err)
				return
			}
		}
	})
	http.ListenAndServe("localhost:5000", nil)
}

type writerFn func(p []byte) (n int, err error)

func (w writerFn) Write(p []byte) (n int, err error) { return w(p) }

func newRTPWriter(w io.Writer, payloadType uint8, payloader rtp.Payloader) writerFn {
	const mtu = 1000
	const clockRate = 9000
	var timestamp time.Time
	var samples = func() (n uint32) {
		now := time.Now()
		n = uint32(math.Round(clockRate * now.Sub(timestamp).Seconds()))
		timestamp = now
		return
	}
	pz := rtp.NewPacketizer(mtu, payloadType, rand.Uint32(),
		payloader, rtp.NewRandomSequencer(), clockRate)
	pktBuffer := make([]byte, mtu)
	return func(p []byte) (n int, err error) {
		for _, pkt := range pz.Packetize(p, samples()) {
			l, err := pkt.MarshalTo(pktBuffer)
			if err != nil {
				return n, err
			}
			if _, err = w.Write(pktBuffer[:l]); err != nil {
				return n, err
			}
			n += l
		}
		return
	}
}
