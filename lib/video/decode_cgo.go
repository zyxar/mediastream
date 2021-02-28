// +build cgo

package video

/*
#include <stdint.h>

void decodeYUY2(
    uint8_t* y,
    uint8_t* cb,
    uint8_t* cr,
    uint8_t* yuy2,
    int width, int height)
{
  const int l = width * height * 2;
  int i, fast = 0, slow = 0;
  for (i = 0; i < l; i += 4)
  {
    y[fast] = yuy2[i];
    cb[slow] = yuy2[i + 1];
    y[fast + 1] = yuy2[i + 2];
    cr[slow] = yuy2[i + 3];
    fast += 2;
    ++slow;
  }
}

void decodeUYVY(
    uint8_t* y,
    uint8_t* cb,
    uint8_t* cr,
    uint8_t* uyvy,
    int width, int height)
{
  const int l = width * height * 2;
  int i, fast = 0, slow = 0;
  for (i = 0; i < l; i += 4)
  {
    cb[slow] = uyvy[i];
    y[fast] = uyvy[i+1];
    cr[slow] = uyvy[i + 2];
    y[fast + 1] = uyvy[i + 3];
    fast += 2;
    ++slow;
  }
}
*/
import "C"

func fillYUY2(y, cb, cr []byte, buf []byte, width, height int) {
	C.decodeYUY2(
		(*C.uchar)(&y[0]),
		(*C.uchar)(&cb[0]),
		(*C.uchar)(&cr[0]),
		(*C.uchar)(&buf[0]),
		C.int(width), C.int(height))
}

func fillUYVY(y, cb, cr []byte, buf []byte, width, height int) {
	C.decodeUYVY(
		(*C.uchar)(&y[0]),
		(*C.uchar)(&cb[0]),
		(*C.uchar)(&cr[0]),
		(*C.uchar)(&buf[0]),
		C.int(width), C.int(height))
}
