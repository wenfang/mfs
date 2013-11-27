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
	fin    chan uint32
}

type Img struct {
	fname string
	fw    *os.File
	fr    *os.File
	wchan chan wdata
	s     *Super
}

func (img *Img) putObj(objLen uint64, src io.Reader) uint32 {
	// get imgPos
	imgPos := img.s.ImgLen
	img.s.ImgLen += objLen + ObjHeadSize + ObjTailSize
	img.fw.Seek(8, 0)
	img.fw.Write(Uint64ToByte(uint64(img.s.ImgLen))[2:])
	// get objId
	objId := img.s.NextObjId
	img.s.NextObjId++
	img.fw.Seek(20, 0)
	img.fw.Write(Uint64ToByte(uint64(img.s.NextObjId))[4:])
	// write idx
	idxOff := img.s.MIdx[objId/MIdxSize-MinMIdx] + uint64(objId%MIdxSize)*IdxEntrySize
	buf := make([]byte, IdxEntrySize)
	copy(buf[0:6], Uint64ToByte(imgPos)[2:])
	copy(buf[8:14], Uint64ToByte(objLen)[2:])
	img.fw.Seek(int64(idxOff), 0)
	img.fw.Write(buf)
	// write obj
	img.fw.Seek(int64(imgPos), 0)
	buf = make([]byte, ObjHeadSize)
	n := copy(buf, []byte("OSTA"))
  n += copy(buf[n:], Uint64ToByte(uint64(objId))[4:])
  n += copy(buf[n:], Uint64ToByte(objLen + ObjHeadSize + ObjTailSize)[4:])
  n += 2
  n += copy(buf[n:], Uint64ToByte(objLen)[4:])
  img.fw.Write(buf)
	io.CopyN(img.fw, src, int64(objLen))

	return objId
}

func (img *Img) wRoutine() {
	defer img.fw.Close()

	for v := range img.wchan {
		switch v.wtype {
		case W_PUTOBJ:
			v.fin <- img.putObj(v.objLen, v.src)
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

func (img *Img) PutObj(objLen uint64, src io.Reader) uint32 {
	fin := make(chan uint32)
	img.wchan <- wdata{W_PUTOBJ, 0, objLen, 0, src, fin}
	return <-fin
}

func (img *Img) GetObj(objId uint64, dst io.Writer) {
	idxOff := img.s.MIdx[objId/MIdxSize-MinMIdx] + uint64(objId%MIdxSize)*IdxEntrySize
	entry := NewIdxEntry(img.fr, idxOff)
	img.fr.Seek(int64(entry.ObjPos+ObjHeadSize), 0)
  io.CopyN(dst, img.fr, int64(entry.ObjLen))
}
