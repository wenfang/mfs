package mfs

import (
	"io"
	"log"
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
	fr    *os.File
	wchan chan wdata
	s     *Super
}

func (img *Img) putObj(objLen uint64, src io.Reader, dst io.WriteSeeker) uint32 {
	objId := img.s.UpdateNextObjId(dst)
	imgPos := img.s.UpdateImgLen(dst, objLen+ObjHeadSize+ObjTailSize)
	// write idx
	var entry IdxEntry
	entry.IdxOff = img.s.GetIdxOff(objId)
	entry.ObjPos = imgPos
	entry.ObjLen = objLen
	entry.Update(dst)
	// write obj
	dst.Seek(int64(imgPos), 0)
	buf := make([]byte, ObjHeadSize)
	n := copy(buf, []byte("OSTA"))
	n += copy(buf[n:], Uint64ToByte(uint64(objId))[4:])
	n += copy(buf[n:], Uint64ToByte(objLen + ObjHeadSize + ObjTailSize)[4:])
	n += 2
	n += copy(buf[n:], Uint64ToByte(objLen)[4:])
	dst.Write(buf)
	io.CopyN(dst, src, int64(objLen))

	return objId
}

func (img *Img) wRoutine() {
	fw, err := os.OpenFile(img.fname, os.O_RDWR, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer fw.Close()

	for v := range img.wchan {
		switch v.wtype {
		case W_PUTOBJ:
			v.fin <- img.putObj(v.objLen, v.src, fw)
		}
	}
}

func NewImg(fname string) *Img {
	img := new(Img)
	img.fname = fname

	var err error
	if img.fr, err = os.Open(fname); err != nil {
		return nil
	}
	img.wchan = make(chan wdata, 512)

	go img.wRoutine()
	return img
}

func (img *Img) LoadSuper() bool {
	img.s = NewSuper(img.fr)
	if img.s == nil {
		return false
	}
	return true
}

func (img *Img) GetObjIdxEntry(objId uint32) *IdxEntry {
	idxOff := img.s.GetIdxOff(objId)
	if idxOff == 0 {
		return nil
	}
	return NewIdxEntry(img.fr, idxOff)
}

func (img *Img) GetObj(entry *IdxEntry, dst io.Writer) {
	img.fr.Seek(int64(entry.ObjPos+ObjHeadSize), 0)
	io.CopyN(dst, img.fr, int64(entry.ObjLen))
}

func (img *Img) PutObj(objLen uint64, src io.Reader) uint32 {
	fin := make(chan uint32)
	img.wchan <- wdata{W_PUTOBJ, 0, objLen, 0, src, fin}
	return <-fin
}
