package mfs

import (
	"os"
)

const (
	SuperSize    = 32 * 1024
	IdxEntrySize = 16
	MinMIdx      = 3
	MaxMIdx      = 4096 - 1
)

type Super struct {
	MIdxSizeBit uint16
	SiteGrp     uint32
	ImgLen      uint64
	ImgSize     uint64
	NextObjId   uint32
	MIdx        [MaxMIdx - MinMIdx + 1]uint64
}

func NewSuper(buf []byte) *Super {
	if len(buf) < SuperSize {
		return nil
	}

	if string(buf[0:3]) != "MFS" {
		return nil
	}
	res := new(Super)

	res.MIdxSizeBit = uint16(byteTouint64(buf[3:4]))
	res.SiteGrp = uint32(byteTouint64(buf[4:8]))
	res.ImgLen = byteTouint64(buf[8:14])
	res.ImgSize = byteTouint64(buf[14:20])
	if res.ImgLen > res.ImgSize {
		return nil
	}

	if res.NextObjId = uint32(byteTouint64(buf[20:24])); res.NextObjId < MinMIdx*(1<<res.MIdxSizeBit) {
		return nil
	}

	for i, pos := uint32(MinMIdx), 24; i <= res.NextObjId/(1<<res.MIdxSizeBit); i, pos = i+1, pos+8 {
		res.MIdx[i-MinMIdx] = byteTouint64(buf[pos : pos+8])
	}

	return res
}

func NewSuperFromFile(f *os.File) *Super {
	buf := make([]byte, SuperSize)
	if _, err := f.Read(buf); err != nil {
		return nil
	}
	return NewSuper(buf)
}

func (s *Super) GetIdxEntryOff(objId uint32) uint64 {
	if objId >= s.NextObjId || objId < MinMIdx*(1<<s.MIdxSizeBit) {
		return 0
	}
	return s.MIdx[objId/(1<<s.MIdxSizeBit)] + uint64(objId%(1<<s.MIdxSizeBit))*IdxEntrySize
}

func (s *Super) UpdateImgLen(f *os.File, ObjSize uint64) {
	s.ImgLen += ObjSize
	f.Seek(8, 0)
	f.Write(uint64Tobyte(s.ImgLen)[2:])
}

func (s *Super) UpdateNextObjId(f *os.File) {
	s.NextObjId++
	f.Seek(20, 0)
	f.Write(uint64Tobyte(uint64(s.NextObjId))[4:])
}

func (s *Super) Flush(f *os.File) {
	buf := make([]byte, SuperSize)
	n := copy(buf, []byte("MFS"))
	n += copy(buf[n:], uint64Tobyte(uint64(s.MIdxSizeBit))[7:])
	n += copy(buf[n:], uint64Tobyte(uint64(s.SiteGrp))[4:])
	n += copy(buf[n:], uint64Tobyte(s.ImgLen)[2:])
	n += copy(buf[n:], uint64Tobyte(s.ImgSize)[2:])
	n += copy(buf[n:], uint64Tobyte(uint64(s.NextObjId))[4:])
	for _, v := range s.MIdx {
		n += copy(buf[n:], uint64Tobyte(v))
	}
	f.Seek(0, 0)
	f.Write(buf)
}
