package mfs

import (
	"testing"
)

var buf []byte

func init() {
	buf = make([]byte, SuperSize)
	buf[0] = 'M'
	buf[1] = 'J'
	buf[2] = 'F'
	buf[3] = 'S'
	buf[21] = 0x30
	buf[30] = 0x80
}

func TestNewSuper(t *testing.T) {
	s := newSuper(buf)
	if s == nil {
		t.Fatal("Create Error")
	}
}
