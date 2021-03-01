// +build darwin

package main

import (
	"flag"
	"fmt"
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
	"github.com/zyxar/mediastream/lib/format"
	"github.com/zyxar/mediastream/lib/video"

	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
)

var (
	selectedFormat    = flag.String("format", "NV12", "set pixel format")
	selectedFrameRate = flag.Float64("framerate", 30, "set frame rate")
	selectedOut       = flag.String("out", "", "set output file name")
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

	type encoder func(w io.Writer, buf []byte) error
	var imageBuffer = make([]byte, s.BufferSize())
	var encode = func(w io.Writer, enc encoder) error {
		if _, err := s.ReadVideoFrame(imageBuffer); err != nil {
			return err
		}
		return enc(w, imageBuffer)
	}

	if *selectedOut != "" {
		codec, err := openh264.NewEncoder(p.Width, p.Height, 500_000, p.FrameRate)
		if err != nil {
			log.Fatal(err)
		}

		var writer io.Writer
		uri, err := url.ParseRequestURI(*selectedOut)
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
			writer = conn
		default:
			file, err := os.Create(*selectedOut)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
			writer = file
		}

		const mtu = 1000
		const payloadType = 125
		const clockRate = 9000
		packetizer := rtp.NewPacketizer(mtu, uint8(payloadType), rand.Uint32(),
			&codecs.H264Payloader{}, rtp.NewRandomSequencer(), clockRate)

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGABRT, syscall.SIGTERM, syscall.SIGHUP)
		defer signal.Stop(sig)

		var frameBuffer = make([]byte, s.BufferSize())
		var pktBuffer = make([]byte, mtu)
		sampler := &sampler{clockRate, time.Now()}
		enc := func(w io.Writer, buf []byte) error {
			img, err := video.DecodeToYUV420(p.PixelFormat, buf, p.Width, p.Height)
			if err != nil {
				return err
			}
			l, err := codec.EncodeFrame(frameBuffer, img)
			if l > 0 {
				for _, pkt := range packetizer.Packetize(frameBuffer[:l], sampler.Samples()) {
					n, err := pkt.MarshalTo(pktBuffer)
					if err != nil {
						log.Println(err)
						break
					}
					if _, err = w.Write(pktBuffer[:n]); err != nil {
						return err
					}
				}
			}
			return err
		}
		for {
			select {
			case <-sig:
				return
			default:
				if err = encode(writer, enc); err != nil {
					log.Println(err)
					return
				}
			}
		}
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mimeWriter := multipart.NewWriter(w)
		contentType := fmt.Sprintf("multipart/x-mixed-replace;boundary=%s", mimeWriter.Boundary())
		w.Header().Add("Content-Type", contentType)
		partHeader := make(textproto.MIMEHeader)
		partHeader.Add("Content-Type", "image/jpeg")
		for {
			partWriter, err := mimeWriter.CreatePart(partHeader)
			if err != nil {
				log.Println(err)
				return
			}
			err = encode(partWriter, func(w io.Writer, buf []byte) error {
				img, err := video.Decode(p.PixelFormat, buf, p.Width, p.Height)
				if err != nil {
					return err
				}
				return jpeg.Encode(w, img, nil)
			})
			if err != nil {
				log.Println(err)
				return
			}
		}
	})
	http.ListenAndServe("localhost:5000", nil)
}

type sampler struct {
	clockRate float64
	timestamp time.Time
}

func (s *sampler) Samples() (n uint32) {
	now := time.Now()
	n = uint32(math.Round(s.clockRate * now.Sub(s.timestamp).Seconds()))
	s.timestamp = now
	return
}
