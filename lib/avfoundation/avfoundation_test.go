package avfoundation

import (
	"testing"

	"github.com/zyxar/mediastream/lib/format"
	"github.com/zyxar/mediastream/lib/video"
)

func TestSession(t *testing.T) {
	s, err := NewSession(Property{PixelFormat: format.YUY2, FrameRate: 30})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	p := s.Property()
	t.Logf("%#v\n", s.Property())
	buf := make([]byte, s.BufferSize())
	for i := 0; i < 10; i++ {
		if i, err := s.ReadVideoFrame(buf); err != nil {
			t.Error(err)
		} else if i != s.BufferSize() {
			t.Error("size mismatch")
		}
		_, err := video.Decode(p.PixelFormat, buf, p.Width, p.Height)
		if err != nil {
			t.Error(err)
		}
	}
}
