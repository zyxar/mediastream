// +build !cgo

package video

func fillYUY2(y, cb, cr []byte, buf []byte, width, height int) {
	fi := width * height * 2
	fast := 0
	slow := 0
	for i := 0; i < fi; i += 4 {
		y[fast] = buf[i]
		cb[slow] = buf[i+1]
		y[fast+1] = buf[i+2]
		cr[slow] = buf[i+3]
		fast += 2
		slow++
	}
}

func fillUYVY(y, cb, cr []byte, buf []byte, width, height int) {
	fi := width * height * 2
	fast := 0
	slow := 0
	for i := 0; i < fi; i += 4 {
		cb[slow] = buf[i]
		y[fast] = buf[i+1]
		cr[slow] = buf[i+2]
		y[fast+1] = buf[i+3]
		fast += 2
		slow++
	}
}
