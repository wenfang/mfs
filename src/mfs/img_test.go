package mfs

import (
	"bytes"
	"testing"
)

func TestImg(t *testing.T) {
	img, err := NewImg("img")
	if err != nil {
		t.Fatal("NewImg Error")
	}

	buf := []byte{'a', 'b', 'c', 'd', 'e', 'f'}
	b := bytes.NewBuffer(nil)
	id, _ := img.Put(3, 10, bytes.NewReader(buf))
	img.Get(id, b)
	if b.String() != "abc" {
		t.Fatal(b.String())
	}
  objLen, _ := img.GetObjLen(id)
  if objLen != 3 {
    t.Fatal(objLen)
  }

	b.Truncate(0)
	id, _ = img.Put(5, 20, bytes.NewReader(buf))
	img.Get(id, b)
	if b.String() != "abcde" {
		t.Fatal(b.String())
	}
  objLen, _ = img.GetObjLen(id)
  if objLen != 5 {
    t.Fatal(objLen)
  }

	b.Truncate(0)
	id, _ = img.Put(6, 30, bytes.NewReader(buf))
	img.Get(id, b)
	if b.String() != "abcdef" {
		t.Fatal(b.String())
	}
  objLen, _ = img.GetObjLen(id)
  if objLen != 6 {
    t.Fatal(objLen)
  }
}
