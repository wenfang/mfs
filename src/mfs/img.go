package mfs

import (
	"io"
	"os"
)

const (
	W_PUTOBJ = iota
	W_UPDATEOBJ
	W_DELOBJ
)

type wdata struct {
	wtype  uint16
	objId  uint32
	objLen uint64
	objOff uint64
	src    io.Reader
	fin    chan string
}

type Img struct {
	fname string
	fw    *os.File
	fr    *os.File
	wchan chan wdata
	s     *Super
}

func (img *Img) putObj(objLen uint64, src io.Reader) {
  // get objId
  objId := img.s.NextObjId
  img.s.NextObjId++
  // write idx
  idxOff := img.s.MIdx[objId/MIdxSize] + uint64(objId%MIdxSize)*IdxEntrySize
  buf := make([]byte, IdxEntrySize)
  copy(buf[8:14], Uint64ToByte(objLen)[2:])
  img.fw.Seek(int64(idxOff), 0)
  img.fw.Write(buf)
}

func (img *Img) wRoutine() {
	defer img.fw.Close()

	for v := range img.wchan {
		switch v.wtype {
		case W_PUTOBJ:
			img.putObj(v.objLen, v.src)
		}
	}
}

func NewImg(fname string) *Img {
	var err error
	img := new(Img)
	img.fname = fname

	if img.fr, err = os.Open(fname); err != nil {
		return nil
	}
	if img.fw, err = os.OpenFile(fname, os.O_RDWR, 0666); err != nil {
		return nil
	}
	img.wchan = make(chan wdata, 512)

	go img.wRoutine()

	return img
}

func (img *Img) LoadSuper() bool {
	img.fr.Seek(0, 0)
	if img.s = NewSuper(img.fr); img.s == nil {
		return false
	}
	return true
}

func (img *Img) PutObj(objLen uint64, src io.Reader) {
	img.wchan <- wdata{W_PUTOBJ, 0, objLen, 0, src, nil}
}

func (img *Img) UpdateObj(objId uint64, objOff, objLen uint64, src io.Reader) {
}
