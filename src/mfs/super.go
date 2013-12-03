package mfs

import (
	"errors"
	"io"
)

const (
	SuperBlockSize = 4096
	MIdxNum        = 4096
	MIdxBlockSize  = 8 * MIdxNum
	BlockAlign     = 1024
	MIdxSize       = 1 << 20
)

type Super struct {
	SiteGrp   uint32
	ImgLen    uint64
	ImgSize   uint64
	NextObjId uint32
	MIdx      [MIdxNum]uint64
}

var (
	SESeek        = errors.New("Super Seek Error")
	SERead        = errors.New("Super Read Error")
	SEMagic       = errors.New("Super Magic Error")
	SEImgLen      = errors.New("Super Img Length Error")
	SEObjId       = errors.New("Super ObjId Error")
	SEIdxOff      = errors.New("Super Idx Offset Error")
	SEImgOver     = errors.New("Super Img Overflow")
	SEImgLenSeek  = errors.New("Super Img Len Seek Error")
	SEImgLenWrite = errors.New("Super Img Len Write Error")
	SEObjIdOver   = errors.New("Super ObjId Overflow")
	SEObjIdSeek   = errors.New("Super ObjId Seek Error")
	SEObjIdWrite  = errors.New("Super ObjId Write Error")
	SEMidxAlloc   = errors.New("Super Midx Alloc Error")
	SEMidxSeek    = errors.New("Super Midx Seek Error")
	SEMidxWrite   = errors.New("Super Midx Write Error")
)

func NewSuper(f io.ReadSeeker) (*Super, error) {
	if _, err := f.Seek(0, 0); err != nil {
		return nil, SESeek
	}

	buf := make([]byte, SuperBlockSize+MIdxBlockSize)
	if _, err := f.Read(buf); err != nil {
		return nil, SERead
	}

	if string(buf[0:4]) != "MJFS" {
		return nil, SEMagic
	}

	s := new(Super)
	s.SiteGrp = uint32(ByteToUint64(buf[4:8]))
	s.ImgLen = ByteToUint64(buf[8:14])
	s.ImgSize = ByteToUint64(buf[14:20])
	if s.ImgLen > s.ImgSize {
		return nil, SEImgLen
	}
	s.NextObjId = uint32(ByteToUint64(buf[20:24]))

	for i := uint32(0); i <= s.NextObjId/MIdxSize; i = i + 1 {
		s.MIdx[i] = ByteToUint64(buf[SuperBlockSize+i*8 : SuperBlockSize+i*8+8])
	}
	return s, nil
}

// 根据对象ID返回对象索引位置，如果对象ID不合法，返回0
func (s *Super) GetIdxOff(objId uint32) (int64, error) {
	if objId > s.NextObjId {
		return 0, SEIdxOff
	}
	return int64(s.MIdx[objId/MIdxSize]) + int64(objId%MIdxSize)*IdxSize, nil
}

// 将ImgLen写入f
func (s *Super) StoreImgLen(f io.WriteSeeker) error {
	if _, err := f.Seek(8, 0); err != nil {
		return SEImgLenSeek
	}
	if _, err := f.Write(Uint64ToByte(s.ImgLen)[2:]); err != nil {
		return SEImgLenWrite
	}
	return nil
}

// 根据当前ImgLen的值计算新对象的存放位置，及存放后ImgLen的值
func (s *Super) NewImgLen(size uint64) (uint64, uint64) {
	var extra uint64
	if size%BlockAlign != 0 {
		extra = 1
	}
	return s.ImgLen, s.ImgLen + (size/BlockAlign+extra)*BlockAlign
}

// 将NextObjId的值保存在f中
func (s *Super) StoreNextObjId(f io.WriteSeeker) error {
	if _, err := f.Seek(20, 0); err != nil {
		return SEObjIdSeek
	}
	if _, err := f.Write(Uint64ToByte(uint64(s.NextObjId))[4:]); err != nil {
		return SEObjIdWrite
	}

	if s.NextObjId%MIdxSize == 0 {
		midx := s.NextObjId / MIdxSize
		s.MIdx[midx], s.ImgLen = s.NewImgLen(MIdxSize * IdxSize)
		if err := s.StoreImgLen(f); err != nil {
			return SEMidxAlloc
		}
		if _, err := f.Seek(int64(SuperBlockSize+midx*8), 0); err != nil {
			return SEMidxSeek
		}
		if _, err := f.Write(Uint64ToByte(s.MIdx[midx])); err != nil {
			return SEMidxWrite
		}
	}
	return nil
}
