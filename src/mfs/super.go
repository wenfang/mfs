package mfs

import (
	"os"
)

const (
	SuperSize    = 32 * 1024
	IdxEntrySize = 16
	FstMainIdx   = 3
	MaxMainIdx   = 4096 - 1
)

type Super struct {
	MainIdxSize uint32
	Id          uint32
	DatOff      uint64
	DatSize     uint64
	IdxNum      uint32
	IdxMap      [MaxMainIdx - FstMainIdx]uint64
}

func byteTouint64(buf []byte) uint64 {
	var res uint64
	for i, v := range buf {
		if i >= 8 {
			break
		}
		res <<= 8
		res += uint64(v)
	}
	return res
}

func uint64Tobyte(v uint64) []byte {
	buf := make([]byte, 8)
	for i := 7; i >= 0; i-- {
    buf[i] = byte(v)
    if v >>= 8; v == 0 {
      break
    }
	}
	return buf
}

func NewSuper(buf []byte) *Super {
	if len(buf) < SuperSize {
		return nil
	}

	if buf[0] != 'M' || buf[1] != 'F' || buf[2] != 'S' {
		return nil
	}
	res := new(Super)

	res.MainIdxSize = 1 << byteTouint64(buf[3:4])
	res.Id = uint32(byteTouint64(buf[4:8]))
	res.DatOff = byteTouint64(buf[8:14])
	res.DatSize = byteTouint64(buf[14:20])
	if res.DatOff < res.DatSize {
		return nil
	}

	res.IdxNum = uint32(byteTouint64(buf[20:24]))
	if res.IdxNum < FstMainIdx*res.MainIdxSize {
		return nil
	}

	for i, pos := uint32(FstMainIdx), 24; i < res.IdxNum/res.MainIdxSize; i, pos = i+1, pos+4 {
		res.IdxMap[i-FstMainIdx] = byteTouint64(buf[pos : pos+4])
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

func (s *Super) UpdateDatOff(f *os.File) {
}

func (s *Super) UpdateIdxNum(f *os.File) {
}

func (s *Super) GetIdxOff(idx uint32) uint64 {
	return uint64(idx * IdxEntrySize)
}
