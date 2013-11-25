package mfs

import (
	"testing"
)

func TestNewMeta(t *testing.T) {
	buf := make([]byte, SuperSize)
	buf[0] = 'M'
	buf[1] = 'F'
	buf[2] = 'S'
	buf[3] = 20
	buf[21] = 0x30
	m := NewSuper(buf)
	if m == nil {
		t.Fatal("Create Error")
	}
}

func TestToByte(t *testing.T) {
  t.Fatal(uint64Tobyte(1232002342))
}
