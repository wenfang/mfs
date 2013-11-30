package mfs

import (
	"errors"
	"io"
)

const (
	SuperSize = 32 * 1024
	MinMIdx   = 3
	MaxMIdx   = 4096 - 1
	MIdxSize  = 1 << 20
)

type Super struct {
	SiteGrp   uint32
	ImgLen    uint64
	ImgSize   uint64
	NextObjId uint32
	MIdx      [MaxMIdx - MinMIdx + 1]uint64
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

	buf := make([]byte, SuperSize)
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
	if s.NextObjId < MinMIdx*MIdxSize {
		return nil, SEObjId
	}
	for i, pos := uint32(MinMIdx), 24; i <= s.NextObjId/MIdxSize; i, pos = i+1, pos+8 {
		s.MIdx[i-MinMIdx] = ByteToUint64(buf[pos : pos+8])
	}
	return s, nil
}

// 根据对象ID返回对象索引位置，如果对象ID不合法，返回0
func (s *Super) GetIdxOff(objId uint32) (int64, error) {
	if objId < MinMIdx*MIdxSize || objId >= s.NextObjId {
		return 0, SEIdxOff
	}
	return int64(s.MIdx[objId/MIdxSize-MinMIdx]) + int64(objId%MIdxSize)*IdxSize, nil
}

func (s *Super) UpdateImgLen(f io.WriteSeeker, objSize uint64) (uint64, error) {
	if s.ImgLen+objSize >= s.ImgSize {
		return 0, SEImgOver
	}

	res := s.ImgLen
	s.ImgLen += objSize
	if _, err := f.Seek(8, 0); err != nil {
		return 0, SEImgLenSeek
	}
	if _, err := f.Write(Uint64ToByte(s.ImgLen)[2:]); err != nil {
		return 0, SEImgLenWrite
	}
	return res, nil
}

// NextObjId加1，返回老的NextObjId，必要时扩展MIdx，出错返回0
func (s *Super) UpdateNextObjId(f io.WriteSeeker) (uint32, error) {
	if s.NextObjId > 0xFFFFFFF0 {
		return 0, SEObjIdOver
	}

	var err error

	res := s.NextObjId
	s.NextObjId++
	if _, err = f.Seek(20, 0); err != nil {
		return 0, SEObjIdSeek
	}
	if _, err = f.Write(Uint64ToByte(uint64(s.NextObjId))[4:]); err != nil {
		return 0, SEObjIdWrite
	}

	if s.NextObjId%MIdxSize == 0 {
		midx := s.NextObjId/MIdxSize - MinMIdx
		if s.MIdx[midx], err = s.UpdateImgLen(f, MIdxSize*IdxSize); err != nil {
			return 0, SEMidxAlloc
		}
		if _, err = f.Seek(int64(24+midx*8), 0); err != nil {
			return 0, SEMidxSeek
		}
		if _, err = f.Write(Uint64ToByte(s.MIdx[midx])); err != nil {
			return 0, SEMidxWrite
		}
	}
	return res, nil
}
