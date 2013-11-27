package mfs

import (
	"os"
)

const (
	SuperSize    = 32 * 1024
	IdxEntrySize = 16
	MinMIdx      = 3
	MaxMIdx      = 4096 - 1
	MIdxSize     = 1 << 20
)

type Super struct {
	SiteGrp   uint32
	ImgLen    uint64
	ImgSize   uint64
	NextObjId uint32
	MIdx      [MaxMIdx - MinMIdx + 1]uint64
}

func newSuper(buf []byte) *Super {
	if len(buf) < SuperSize {
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

func NewSuper(f *os.File) *Super {
	buf := make([]byte, SuperSize)
	if _, err := f.Read(buf); err != nil {
		return nil
	}
	return newSuper(buf)
}

func (s *Super) UpdateImgLen(f *os.File, ObjSize uint64) {
	s.ImgLen += ObjSize
	f.Seek(8, 0)
	f.Write(Uint64ToByte(s.ImgLen)[2:])
}

func (s *Super) UpdateNextObjId(f *os.File) {
	s.NextObjId++
	f.Seek(20, 0)
	f.Write(Uint64ToByte(uint64(s.NextObjId))[4:])
}
