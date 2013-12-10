package mfs

import (
	"bytes"
	"testing"
)

var buf = []byte{'a', 'b', 'c', 'd', 'e', 'f', '1', '4', '6', '8', '2'}

var Table = []struct {
	objLen    uint64
	objSize   uint64
	updateLen uint64
}{
	{3, 10, 5},
	{5, 20, 3},
	{6, 30, 2},
	{10, 1, 9},
}

func TestImg(t *testing.T) {
	img, err := NewImg("img")
	if err != nil {
		t.Fatal("NewImg Error")
	}

	for _, v := range Table {
		var b bytes.Buffer
		id, err := img.Put(v.objLen, v.objSize, bytes.NewReader(buf))
		if err != nil {
			t.Fatal(err)
		}

		err = img.Get(id, &b)
		if err != nil {
			t.Fatal(err)
		}
		if b.String() != string(buf[:v.objLen]) {
			t.Fatal(id, b.String())
		}

		objLen, err := img.GetObjLen(id)
		if err != nil {
			t.Fatal(err)
		}
		if objLen != v.objLen {
			t.Fatal(objLen, v.objLen)
		}

		err = img.Update(id, v.updateLen, bytes.NewReader(buf))
		if err != nil {
			t.Fatal(err)
		}

		b.Truncate(0)
		err = img.Get(id, &b)
		if err != nil {
			t.Fatal(err)
		}
		if b.String() != string(buf[:v.updateLen]) {
			t.Fatal(id, b.String())
		}

		objLen, err = img.GetObjLen(id)
		if err != nil {
			t.Fatal(err)
		}
		if objLen != v.updateLen {
			t.Fatal(objLen, v.updateLen)
		}

		err = img.Del(id)
		if err != nil {
			t.Fatal(err)
		}
	}
}
