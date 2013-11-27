package mfs

import (
	"bytes"
	"testing"
)

func TestImg(t *testing.T) {
	img := NewImg("img")
	if img == nil {
		t.Fatal("NewImg Error")
	}

	if !img.LoadSuper() {
		t.Fatal("Load Super Error")
	}
	buf := []byte{'a', 'b', 'c', 'd', 'e', 'f'}
	b := bytes.NewBuffer(nil)
	id := img.PutObj(3, bytes.NewReader(buf))
	img.GetObj(uint64(id), b)
	if b.String() != "abc" {
		t.Fatal(b.String())
	}

	b.Truncate(0)
	id = img.PutObj(5, bytes.NewReader(buf))
	img.GetObj(uint64(id), b)
	if b.String() != "abcde" {
		t.Fatal(b.String())
	}

	b.Truncate(0)
	id = img.PutObj(6, bytes.NewReader(buf))
	img.GetObj(uint64(id), b)
	if b.String() != "abcdef" {
		t.Fatal(b.String())
	}
}
