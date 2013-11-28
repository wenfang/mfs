package mfs

import (
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

func NewSuper(src io.ReadSeeker) *Super {
	src.Seek(0, 0)
	buf := make([]byte, SuperSize)
	if _, err := src.Read(buf); err != nil {
		return nil
	}

	if string(buf[0:4]) != "MJFS" {
		return nil
	}

	s := new(Super)
	s.SiteGrp = uint32(ByteToUint64(buf[4:8]))
	s.ImgLen = ByteToUint64(buf[8:14])
	s.ImgSize = ByteToUint64(buf[14:20])
	if s.ImgLen > s.ImgSize {
		return nil
	}
	s.NextObjId = uint32(ByteToUint64(buf[20:24]))
	if s.NextObjId < MinMIdx*MIdxSize {
		return nil
	}
	for i, pos := uint32(MinMIdx), 24; i <= s.NextObjId/MIdxSize; i, pos = i+1, pos+8 {
		s.MIdx[i-MinMIdx] = ByteToUint64(buf[pos : pos+8])
	}
	return s
}

// 根据对象ID返回对象索引位置，如果对象ID不合法，返回0
func (s *Super) GetIdxOff(objId uint32) int64 {
	if objId < MinMIdx*MIdxSize || objId >= s.NextObjId {
		return 0
	}
	return int64(s.MIdx[objId/MIdxSize-MinMIdx]) + int64(objId%MIdxSize)*IdxSize
}

func (s *Super) UpdateImgLen(f io.WriteSeeker, objSize uint64) (res uint64) {
	if s.ImgLen+objSize >= s.ImgSize {
		return
	}

	res = s.ImgLen
	s.ImgLen += objSize
	f.Seek(8, 0)
	f.Write(Uint64ToByte(s.ImgLen)[2:])
	return
}

// NextObjId加1，返回老的NextObjId，必要时扩展MIdx，出错返回0
func (s *Super) UpdateNextObjId(f io.WriteSeeker) (res uint32) {
	if s.NextObjId > 0xFFFFFFF0 {
		return
	}

	res = s.NextObjId
	s.NextObjId++
	f.Seek(20, 0)
	f.Write(Uint64ToByte(uint64(s.NextObjId))[4:])

	if s.NextObjId%MIdxSize == 0 {
		midx := s.NextObjId/MIdxSize - MinMIdx
		s.MIdx[midx] = s.UpdateImgLen(f, MIdxSize*IdxSize)
		f.Seek(int64(24+midx*8), 0)
		f.Write(Uint64ToByte(s.MIdx[midx]))
	}
	return
}
