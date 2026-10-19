package tray

import (
	"encoding/binary"
	"testing"
)

// fakeICO builds an .ico with one entry per width; each image's data is
// a single byte equal to its width (256 stored as 0).
func fakeICO(widths ...int) []byte {
	head := make([]byte, 6+16*len(widths))
	binary.LittleEndian.PutUint16(head[2:], 1)
	binary.LittleEndian.PutUint16(head[4:], uint16(len(widths)))
	data := []byte{}
	for i, w := range widths {
		e := 6 + 16*i
		head[e] = byte(w % 256)
		binary.LittleEndian.PutUint32(head[e+8:], 1)
		binary.LittleEndian.PutUint32(head[e+12:], uint32(len(head)+len(data)))
		data = append(data, byte(w%256))
	}
	return append(head, data...)
}

func TestPickIcon(t *testing.T) {
	ico := fakeICO(16, 24, 32, 48, 256)
	cases := map[int]byte{16: 16, 20: 24, 32: 32, 40: 48, 300: 0 /* 256 */}
	for size, want := range cases {
		got := pickIcon(ico, size)
		if len(got) != 1 || got[0] != want {
			t.Errorf("size %d: got %v want %d", size, got, want)
		}
	}
	if pickIcon([]byte("nope"), 16) != nil || pickIcon(nil, 16) != nil {
		t.Error("invalid data should give nil")
	}
	broken := fakeICO(16)
	binary.LittleEndian.PutUint32(broken[6+12:], 9999) // offset past the end
	if pickIcon(broken, 16) != nil {
		t.Error("out-of-range entry should be skipped")
	}
}
