package mfs

import (
	"bytes"
	"errors"
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

var (
	IENewSuper = errors.New("Img New Super Error")
	IEObjDel   = errors.New("Img Obj Deleted")
	IEIdxObj   = errors.New("Img Idx Obj Not Match")
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

func (img *Img) putObj(objLen uint64, src io.Reader, f io.WriteSeeker) uint32 {
	objId, err := img.Sup.UpdateNextObjId(f)
	if err != nil {
		return 0
	}

	imgPos, err := img.Sup.UpdateImgLen(f, objLen+ObjHeadSize+ObjTailSize)
	if err != nil {
		return 0
	}

	var idx Idx
	idx.Offset, _ = img.Sup.GetIdxOff(objId)
	idx.ObjPos = imgPos
	idx.ObjLen = objLen
	idx.Store(f)

	var obj Obj
	obj.Offset = int64(imgPos)
	obj.ObjId = objId
	obj.ObjSize = objLen + ObjHeadSize + ObjTailSize
	obj.ObjLen = objLen
	obj.StoreHead(f)
	obj.StoreData(src, f)

	return objId
}

func (img *Img) delObj(objId uint32, f io.WriteSeeker) uint32 {
	fr := img.pool.Alloc()
	defer img.pool.Free(fr)

	idx, err := img.getIdx(objId, fr)
	if err != nil {
		return 0
	}
	idx.ObjFlag |= 0x1
	idx.Store(f)

	obj, err := img.getObj(idx, fr)
	if err != nil {
		return 1
	}
	obj.ObjFlag |= 0x1
	obj.StoreHead(f)

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
		default:
		}
	}
}

// 创建新的Img结构
func NewImg(ImgName string) (*Img, error) {
	img := new(Img)
	img.ImgName = ImgName
	img.pool = NewFPool(ImgName)
	img.wchan = make(chan wdata, 512)

	fr := img.pool.Alloc()
	defer img.pool.Free(fr)

	var err error
	img.Sup, err = NewSuper(fr)
	if err != nil {
		return nil, IENewSuper
	}

	go img.wRoutine()
	return img, nil
}

func (img *Img) getIdx(objId uint32, fr *os.File) (*Idx, error) {
	offset, err := img.Sup.GetIdxOff(objId)
	if err != nil {
		return nil, err
	}

	idx, err := NewIdx(fr, offset)
	if err != nil {
		return nil, err
	}
	return idx, nil
}

func (img *Img) getObj(idx *Idx, fr *os.File) (*Obj, error) {
	obj, err := NewObj(fr, int64(idx.ObjPos))
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// 获得objId所对应的对象的长度，排除已删除对象，对象未找到，返回0
func (img *Img) GetObjLen(objId uint32) (uint64, error) {
	fr := img.pool.Alloc()
	defer img.pool.Free(fr)

	idx, err := img.getIdx(objId, fr)
	if err != nil {
		return 0, err
	}
	if idx.ObjFlag&0x1 == 0x1 {
		return 0, IEObjDel
	}

	return idx.ObjLen, nil
}

func (img *Img) Get(objId uint32, c io.Writer) error {
	fr := img.pool.Alloc()
	defer img.pool.Free(fr)

	idx, err := img.getIdx(objId, fr)
	if err != nil {
		return err
	}

	obj, err := img.getObj(idx, fr)
	if err != nil {
		return err
	}

	if idx.ObjType != obj.ObjType || idx.ObjLen != obj.ObjLen || idx.ObjFlag != obj.ObjFlag {
		return IEIdxObj
	}
	if idx.ObjFlag&0x1 == 0x1 {
		return IEObjDel
	}

	return obj.Retrive(fr, c)
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

func (img *Img) Update(objId uint32, offset, uptLen uint64) {
    fin := make(chan uint32)
    img.wchan <- wdata{UPDATEOBJ, objId,
}
