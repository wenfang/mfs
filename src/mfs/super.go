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

func NewSuper(f io.ReadSeeker) *Super {
	f.Seek(0, 0)
	buf := make([]byte, SuperSize)
	if _, err := f.Read(buf); err != nil {
		return nil
	}

	if string(buf[0:4]) != "MJFS" {
		return nil
	}

	s := new(Super)
	s.SiteGrp = uint32(ByteToUint64(buf[4:8]))
	if s.ImgLen, s.ImgSize = ByteToUint64(buf[8:14]), ByteToUint64(buf[14:20]); s.ImgLen > s.ImgSize {
		return nil
	}
	if s.NextObjId = uint32(ByteToUint64(buf[20:24])); s.NextObjId < MinMIdx*MIdxSize {
		return nil
	}
	for i, pos := uint32(MinMIdx), 24; i <= s.NextObjId/MIdxSize; i, pos = i+1, pos+8 {
		s.MIdx[i-MinMIdx] = ByteToUint64(buf[pos : pos+8])
	}
	return s
}

// 根据对象ID返回对象索引位置，如果对象ID不合法，返回0
func (s *Super) GetIdxOff(objId uint32) uint64 {
	if objId < MinMIdx*MIdxSize || objId >= s.NextObjId {
		return 0
	}
	return s.MIdx[objId/MIdxSize-MinMIdx] + uint64(objId%MIdxSize)*IdxEntrySize
}

func (s *Super) UpdateImgLen(f io.WriteSeeker, objSize uint64) uint64 {
	oldImgLen := s.ImgLen
	s.ImgLen += objSize
	f.Seek(8, 0)
	f.Write(Uint64ToByte(s.ImgLen)[2:])
	return oldImgLen
}

func (s *Super) updateMIdx(f io.WriteSeeker, mIdx uint32) {
	f.Seek(int64(24+mIdx*8), 0)
	f.Write(Uint64ToByte(s.MIdx[mIdx]))
}

func (s *Super) UpdateNextObjId(f io.WriteSeeker) uint32 {
	oldObjId := s.NextObjId
	s.NextObjId++
	f.Seek(20, 0)
	f.Write(Uint64ToByte(uint64(s.NextObjId))[4:])

	if s.NextObjId%MIdxSize == 0 {
		s.MIdx[s.NextObjId/MIdxSize-MinMIdx] = s.UpdateImgLen(f, MIdxSize*IdxEntrySize)
		s.updateMIdx(f, s.NextObjId/MIdxSize-MinMIdx)
	}
	return oldObjId
}
