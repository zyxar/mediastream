// +build darwin

package main

import (
	"flag"
	"fmt"
	"image/jpeg"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/zyxar/mediastream/lib/codec/openh264"

	"github.com/zyxar/mediastream/lib/avfoundation"
	"github.com/zyxar/mediastream/lib/format"
	"github.com/zyxar/mediastream/lib/image"
)

var (
	selectedFormat    = flag.String("format", "NV12", "set pixel format")
	selectedFrameRate = flag.Float64("framerate", 30, "set frame rate")
	selectedOut       = flag.String("out", "", "set output file name")
)

func main() {
	flag.Parse()

	var pixelFormat = format.PixelFormat(strings.ToUpper(*selectedFormat))
	s, err := avfoundation.NewSession(avfoundation.Property{PixelFormat: pixelFormat, FrameRate: *selectedFrameRate})
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	p := s.Property()

	type encoder func(w io.Writer, i image.Image) error
	var buf = make([]byte, s.BufferSize())
	var encode = func(w io.Writer, enc encoder) error {
		if _, err := s.ReadVideoFrame(buf); err != nil {
			return err
		}
		i, err := image.Decode(p.PixelFormat, buf, p.Width, p.Height)
		if err != nil {
			return err
		}
		return enc(w, i)
	}

	if *selectedOut != "" {
		codec, err := openh264.NewEncoder(p.Width, p.Height, 500_000, p.FrameRate)
		if err != nil {
			log.Fatal(err)
		}
		file, err := os.Create(*selectedOut)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		s := make(chan os.Signal, 1)
		signal.Notify(s, syscall.SIGINT, syscall.SIGABRT, syscall.SIGTERM, syscall.SIGHUP)
		defer signal.Stop(s)

		buf := make([]byte, p.Width*p.Height)
		enc := func(w io.Writer, i image.Image) error {
			img, _ := image.Convert(i)
			l, err := codec.EncodeFrame(buf, img)
			if l > 0 {
				_, err = w.Write(buf[:l])
			}
			return err
		}
		for {
			select {
			case <-s:
				return
			default:
				if err = encode(file, enc); err != nil {
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
			err = encode(partWriter, func(w io.Writer, i image.Image) error {
				return jpeg.Encode(w, i, nil)
			})
			if err != nil {
				log.Println(err)
				return
			}
		}
	})
	http.ListenAndServe("localhost:5000", nil)
}
