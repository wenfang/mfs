package mfs

import (
	"bytes"
	"io"
	"log"
	"os"
)

const (
	PUTOBJ = iota
	UPDATEOBJ
	DELOBJ
)

const (
	ObjBufLimit = 64 * 1024 * 1024
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
	Sup     *Super
	pool    *FPool
	wchan   chan wdata
}

func (img *Img) putObj(objLen uint64, src io.Reader, dst io.WriteSeeker) uint32 {
	objId, err := img.Sup.UpdateNextObjId(dst)
	if err != nil {
		return 0
	}

	imgPos, err := img.Sup.UpdateImgLen(dst, objLen+ObjHeadSize+ObjTailSize)
	if err != nil {
		return 0
	}

	var idx Idx
	idx.Offset, _ = img.Sup.GetIdxOff(objId)
	idx.ObjPos = imgPos
	idx.ObjLen = objLen
	idx.Store(dst)

	var obj Obj
	obj.Offset = int64(imgPos)
	obj.ObjId = objId
	obj.ObjSize = objLen + ObjHeadSize + ObjTailSize
	obj.ObjLen = objLen
	obj.StoreHead(dst)
	obj.StoreData(src, dst)

	return objId
}

func (img *Img) delObj(objId uint32, dst io.WriteSeeker) uint32 {
	obj := img.GetObj(objId)
	if obj == nil {
		return 1
	}
	obj.ObjFlag |= 0x1
	obj.StoreHead(dst)

	idx := img.getIdx(objId)
	if idx == nil {
		return 0
	}
	idx.ObjFlag |= 0x1
	idx.Store(dst)

	return 2
}

func (img *Img) wRoutine() {
	fw, err := os.OpenFile(img.ImgName, os.O_RDWR, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer fw.Close()

	for v := range img.wchan {
		switch v.wtype {
		case PUTOBJ:
			v.fin <- img.putObj(v.objLen, v.src, fw)
		case DELOBJ:
			v.fin <- img.delObj(v.objId, fw)
		}
	}
}

// 创建新的Img结构
func NewImg(ImgName string) *Img {
	img := new(Img)
	img.ImgName = ImgName
	img.pool = NewFPool(ImgName)
	img.wchan = make(chan wdata, 512)

	fr := img.pool.Alloc()
	defer img.pool.Free(fr)

  var err error
	img.Sup, err = NewSuper(fr)
	if err != nil{
		return nil
	}

	go img.wRoutine()
	return img
}

func (img *Img) getIdx(objId uint32) *Idx {
	offset, err := img.Sup.GetIdxOff(objId)
	if err != nil {
		return nil
	}

	fr := img.pool.Alloc()
	defer img.pool.Free(fr)

	idx, err := NewIdx(fr, offset)
	if err != nil {
		return nil
	}
	return idx
}

func (img *Img) GetObj(objId uint32) *Obj {
	idx := img.getIdx(objId)
	if idx == nil {
		return nil
	}
	if idx.ObjFlag&0x1 == 0x1 {
		return nil
	}

	fr := img.pool.Alloc()
	defer img.pool.Free(fr)

	obj := NewObj(fr, int64(idx.ObjPos))
	if obj == nil {
		return nil
	}

	if idx.ObjType != obj.ObjType || idx.ObjLen != obj.ObjLen || idx.ObjFlag != obj.ObjFlag {
		return nil
	}
	return obj
}

// 获得objId所对应的对象的长度，排除已删除对象，对象未找到，返回0
func (img *Img) GetObjLen(objId uint32) uint64 {
	idx := img.getIdx(objId)
	if idx == nil || idx.ObjFlag&0x1 == 0x1 {
		return 0
	}
	return idx.ObjLen
}

func (img *Img) Get(o *Obj, dst io.Writer) {
	fr := img.pool.Alloc()
	defer img.pool.Free(fr)

	o.Retrive(fr, dst)
}

// 将objLen长度的数据保存在img，返回保存id，保存失败返回0
func (img *Img) Put(objLen uint64, src io.Reader) uint32 {
	var s *bytes.Buffer
	if objLen <= ObjBufLimit {
		s = bytes.NewBuffer(make([]byte, 0, objLen))
		if _, err := io.CopyN(s, src, int64(objLen)); err != nil {
			return 0
		}
	}
	fin := make(chan uint32)
	img.wchan <- wdata{PUTOBJ, 0, objLen, 0, s, fin}
	return <-fin
}

// 删除objId对应的对象
func (img *Img) Del(objId uint32) uint32 {
	fin := make(chan uint32)
	img.wchan <- wdata{DELOBJ, objId, 0, 0, nil, fin}
	return <-fin
}
