package mfs

import (
	"testing"
)

var buf []byte

func init() {
	buf = make([]byte, SuperSize)
	buf[0] = 'M'
	buf[1] = 'F'
	buf[2] = 'S'
	buf[3] = 20
	buf[21] = 0x30
	buf[30] = 0x80
}

func TestNewSuper(t *testing.T) {
	s := NewSuper(buf)
	if s == nil {
		t.Fatal("Create Error")
	}
}
