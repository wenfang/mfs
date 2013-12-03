package mfs

import (
	"os"
	"testing"
)

var f *os.File

func TestInit(t *testing.T) {
	var err error
	if f, err = os.OpenFile("img", os.O_RDWR, 0666); err != nil {
		t.Fatal("Open File Error")
	}
}

func TestNewSuper(t *testing.T) {
	s, err := NewSuper(f)
	if err != nil {
		t.Fatal("Create Error")
	}

	s.UpdateImgLen(f, 1000)
	markImgLen := s.ImgLen
	s, err = NewSuper(f)
	if s.ImgLen != markImgLen {
		t.Fatal(s.ImgLen, markImgLen)
	}
}

func TestNewImgLen(t *testing.T) {
	var s Super
	s.ImgLen = 1024
	t.Fatal(s.NewImgLen(34569234))
}
