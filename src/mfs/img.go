package mfs

import (
	"io"
	"os"
)

type wdata struct {
	imgOff int64
	buf    io.Reader
	datLen int64
	fin    chan string
}

type Img struct {
	Fname string
	fr    *os.File
	fw    *os.File
	wchan chan wdata
	s     *Super
}

func newWriteRoutine(fname string) chan wdata {
	fw, err := os.OpenFile(fname, os.O_RDWR, 0666)
	if err != nil {
		return nil
	}

	wchan := make(chan wdata, 512)
	go func() {
		for v := range wchan {
      fw.Seek(v.imgOff, 0)
      io.CopyN(fw, v.buf, v.datLen)
      v.fin <- "OK"
		}
	}()
	return wchan
}

func NewImg(fname string) *Img {
	var err error
	img := new(Img)
	img.Fname = fname

	if img.fr, err = os.Open(fname); err != nil {
		return nil
	}

	if img.fw, err = os.OpenFile(fname, os.O_RDWR, 0666); err != nil {
		return nil
	}

	return img
}

func (img *Img) InitSuper(SiteGrp uint32) bool {
	fi, err := os.Stat(img.Fname)
	if err != nil {
		return false
	}

	buf := make([]byte, SuperSize)
	n := copy(buf, []byte("MJFS"))
	n += copy(buf[n:], uint64Tobyte(uint64(SiteGrp))[4:])
	n += copy(buf[n:], uint64Tobyte(SuperSize + MIdxSize*IdxEntrySize)[2:])
	n += copy(buf[n:], uint64Tobyte(uint64(fi.Size()))[2:])
	n += copy(buf[n:], uint64Tobyte(uint64(MinMIdx * MIdxSize))[4:])

	img.fw.Seek(0, 0)
	img.fw.Write(buf)
	return true
}

func (img *Img) LoadSuper() bool {
	img.fr.Seek(0, 0)
	if img.s = NewSuper(img.fr); img.s == nil {
		return false
	}
	return true
}
