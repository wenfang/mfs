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
	ImgName string
	p       *FPool
	wchan   chan wdata
	s       *Super
}

func (img *Img) putObj(objLen uint64, src io.Reader, dst io.WriteSeeker) uint32 {
	objId := img.s.UpdateNextObjId(dst)
	imgPos := img.s.UpdateImgLen(dst, objLen+ObjHeadSize+ObjTailSize)
	// write idx
	var idx Idx
	idx.Offset = img.s.GetIdxOff(objId)
	idx.ObjPos = imgPos
	idx.ObjLen = objLen
	idx.Update(dst)
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
	fw, err := os.OpenFile(img.ImgName, os.O_RDWR, 0666)
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

func NewImg(ImgName string) *Img {
	img := new(Img)
	img.ImgName = ImgName
	img.p = NewFPool(ImgName)
	img.wchan = make(chan wdata, 512)

	go img.wRoutine()
	return img
}

func (img *Img) LoadSuper() bool {
	fr := img.p.Alloc()
	defer img.p.Free(fr)

	img.s = NewSuper(fr)
	if img.s == nil {
		return false
	}
	return true
}

func (img *Img) GetIdx(objId uint32) *Idx {
	offset := img.s.GetIdxOff(objId)
	if offset == 0 {
		return nil
	}

	fr := img.p.Alloc()
	defer img.p.Free(fr)

	return NewIdx(fr, offset)
}

func (img *Img) GetObj(objId uint32) *Obj {
	idx := img.GetIdx(objId)
	if idx == nil {
		return nil
	}

	fr := img.p.Alloc()
	defer img.p.Free(fr)

	o := NewObj(fr, int64(idx.ObjPos))
	if o == nil {
		return nil
	}
	if idx.ObjType != o.ObjType || idx.ObjLen != o.ObjLen || idx.ObjFlag != o.ObjFlag {
		return nil
	}
	return o
}

func (img *Img) Get(o *Obj, dst io.Writer) {
	fr := img.p.Alloc()
	defer img.p.Free(fr)

	o.Retrive(fr, dst)
}

func (img *Img) Put(objLen uint64, src io.Reader) uint32 {
	if objLen <= ObjBufferLimit {
	}

	fin := make(chan uint32)
	img.wchan <- wdata{W_PUTOBJ, 0, objLen, 0, src, fin}
	return <-fin
}
