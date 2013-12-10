package mfs

import (
	"os"
	"testing"
)

func TestObjInit(t *testing.T) {
	if _, err := os.Create("obj"); err != nil {
		t.Fatal(err)
	}
}

func TestObj(t *testing.T) {
	f, err := os.OpenFile("obj", os.O_RDWR, 0666)
	if err != nil {
		t.Fatal(err)
	}

	var obj Obj
	obj.Offset = 0
	obj.ObjId = 1
	obj.ObjSize = 100
	obj.ObjType = 0
	obj.ObjLen = 20
	obj.ObjFlag = 30
	err = obj.StoreHead(f)
	if err != nil {
		t.Fatal(err)
	}
}

func TestObjDeInit(t *testing.T) {
	if err := os.Remove("obj"); err != nil {
		t.Fatal(err)
	}
}
