package mfs

import (
	"bytes"
	"errors"
	"fmt"
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

type wrsp struct {
	rsp uint32
	err error
}

type wreq struct {
	wtype   uint16
	objId   uint32
	objLen  uint64
	objSize uint64
	src     io.Reader
	fin     chan wrsp
}

type Img struct {
	ImgName string
	Sup     *Super
	pool    *FPool
	wchan   chan wreq
}

func (img *Img) putObj(objLen, objSize uint64, b io.Reader, f io.WriteSeeker) (uint32, error) {
	if objSize < objLen {
		objSize = objLen
	}

	objId := img.Sup.NextObjId
	imgPos, imgLen := img.Sup.NewImgLen(objSize + ObjHeadSize + ObjTailSize)

	var idx Idx
	var err error
	if idx.Offset, err = img.Sup.GetIdxOff(objId); err != nil {
		return 0, err
	}
	idx.ObjPos = imgPos
	idx.ObjLen = objLen

	var obj Obj
	obj.Offset = int64(imgPos)
	obj.ObjId = objId
	obj.ObjSize = imgLen - imgPos - ObjHeadSize - ObjTailSize
	obj.ObjLen = objLen
	if err = obj.StoreHead(f); err != nil {
		return 0, err
	}
	if err = obj.StoreData(b, f); err != nil {
		return 0, err
	}

	if err = idx.Store(f); err != nil {
		return 0, err
	}

	img.Sup.NextObjId++
	if err := img.Sup.StoreNextObjId(f); err != nil {
		return 0, err
	}

	img.Sup.ImgLen = imgLen
	if err := img.Sup.StoreImgLen(f); err != nil {
		return 0, err
	}

	return objId, nil
}

func (img *Img) delObj(objId uint32, f io.WriteSeeker) (uint32, error) {
	fr := img.pool.Alloc()
	defer img.pool.Free(fr)

	idx, err := img.getIdx(objId, fr)
	if err != nil {
		return 0, err
	}

	obj, err := img.getObj(idx, fr)
	if err != nil {
		return 0, err
	}
	obj.ObjFlag |= 0x1
	if err = obj.StoreHead(f); err != nil {
		return 0, err
	}

	idx.ObjFlag |= 0x1
	if err = idx.Store(f); err != nil {
		return 0, err
	}

	return 1, nil
}

func (img *Img) updateObj(objId uint32, objLen uint64, b io.Reader, f io.WriteSeeker) (uint32, error) {
	fr := img.pool.Alloc()
	defer img.pool.Free(fr)

	idx, err := img.getIdx(objId, fr)
	if err != nil {
		return 0, err
	}
	obj, err := img.getObj(idx, fr)
	if err != nil {
		return 0, err
	}
	if obj.ObjSize < objLen {
		return 0, errors.New("ObjLen too Large")
	}

	obj.ObjLen = objLen
	if err = obj.StoreHead(f); err != nil {
		return 0, err
	}
	if err = obj.StoreData(b, f); err != nil {
		return 0, err
	}

	idx.ObjLen = objLen
	if err = idx.Store(f); err != nil {
		return 0, err
	}
	return obj.ObjId, nil
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
			rsp, err := img.putObj(v.objLen, v.objSize, v.src, fw)
			v.fin <- wrsp{rsp, err}
		case DELOBJ:
			rsp, err := img.delObj(v.objId, fw)
			v.fin <- wrsp{rsp, err}
		case UPDATEOBJ:
			rsp, err := img.updateObj(v.objId, v.objLen, v.src, fw)
			v.fin <- wrsp{rsp, err}
		}
	}
}

// 创建新的Img结构
func NewImg(ImgName string) (*Img, error) {
	img := new(Img)
	img.ImgName = ImgName
	img.pool = NewFPool(ImgName)
	img.wchan = make(chan wreq, 512)

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
	if idx.ObjFlag&0x1 == 0x1 {
		return nil, IEObjDel
	}

	return idx, nil
}

func (img *Img) getObj(idx *Idx, fr *os.File) (*Obj, error) {
	obj, err := NewObj(fr, int64(idx.ObjPos))
	if err != nil {
		return nil, err
	}

	if idx.ObjType != obj.ObjType || idx.ObjLen != obj.ObjLen || idx.ObjFlag != obj.ObjFlag {
		fmt.Println(idx, obj)
		return nil, IEIdxObj
	}
	return obj, nil
}

func cTob(objLen uint64, c io.Reader) (io.Reader, error) {
	if objLen <= ObjBufLimit {
		b := bytes.NewBuffer(make([]byte, 0, objLen))
		if _, err := io.CopyN(b, c, int64(objLen)); err != nil {
			return nil, err
		}
		return b, nil
	}
	return nil, errors.New("Img Data is too Large")
}

// 获得objId所对应的对象的长度，排除已删除对象，对象未找到，返回0
func (img *Img) GetObjLen(objId uint32) (uint64, error) {
	fr := img.pool.Alloc()
	defer img.pool.Free(fr)

	idx, err := img.getIdx(objId, fr)
	if err != nil {
		return 0, err
	}
	return idx.ObjLen, nil
}

// 获得objId所对应的对象的内容,传输到c中
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

	return obj.Retrive(fr, c)
}

// 将objLen长度的数据保存在img，返回保存id，保存失败返回0
func (img *Img) Put(objLen, objSize uint64, c io.Reader) (uint32, error) {
	b, err := cTob(objLen, c)
	if err != nil {
		return 0, err
	}

	fin := make(chan wrsp)
	img.wchan <- wreq{PUTOBJ, 0, objLen, objSize, b, fin}
	res := <-fin
	return res.rsp, res.err
}

// 删除objId对应的对象
func (img *Img) Del(objId uint32) error {
	fin := make(chan wrsp)
	img.wchan <- wreq{DELOBJ, objId, 0, 0, nil, fin}
	res := <-fin
	return res.err
}

// 更新objId所对应的对象
func (img *Img) Update(objId uint32, objLen uint64, c io.Reader) error {
	b, err := cTob(objLen, c)
	if err != nil {
		return err
	}

	fin := make(chan wrsp)
	img.wchan <- wreq{UPDATEOBJ, objId, objLen, 0, b, fin}
	res := <-fin
	return res.err
}
