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

const (
	ObjBufferLimit = 64 * 1024 * 1024
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
	entry.Offset = img.s.GetIdxEntryOff(objId)
	entry.ObjPos = imgPos
	entry.ObjLen = objLen
	entry.Update(dst)
	// write obj
	var o Obj
	o.Offset = int64(imgPos)
	o.ObjId = objId
	o.ObjSize = objLen + ObjHeadSize + ObjTailSize
	o.ObjLen = objLen
	o.Store(src, dst)

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

func (img *Img) GetIdxEntry(objId uint32) *IdxEntry {
	offset := img.s.GetIdxEntryOff(objId)
	if offset == 0 {
		return nil
	}
	return NewIdxEntry(img.fr, offset)
}

func (img *Img) GetObj(objId uint32) *Obj {
	entry := img.GetIdxEntry(objId)
	if entry == nil {
		return nil
	}

	o := NewObj(img.fr, int64(entry.ObjPos))
	if o == nil {
		return nil
	}
	if entry.ObjType != o.ObjType || entry.ObjLen != o.ObjLen || entry.ObjFlag != o.ObjFlag {
		return nil
	}
	return o
}

func (img *Img) Get(o *Obj, dst io.Writer) {
	o.Retrive(img.fr, dst)
}

func (img *Img) Put(objLen uint64, src io.Reader) uint32 {
	if objLen <= ObjBufferLimit {
	}

	fin := make(chan uint32)
	img.wchan <- wdata{W_PUTOBJ, 0, objLen, 0, src, fin}
	return <-fin
}
