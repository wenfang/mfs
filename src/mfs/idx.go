package mfs

import (
	"io"
)

const (
	IdxSize = 16
)

type Idx struct {
	Offset  int64
	ObjPos  uint64
	ObjType uint16
	ObjLen  uint64
	ObjFlag uint16
}

func NewIdx(src io.ReadSeeker, Offset int64) *Idx {
	src.Seek(Offset, 0)
	buf := make([]byte, IdxSize)
	if _, err := src.Read(buf); err != nil {
		return nil
	}

	idx := new(Idx)
	idx.Offset = Offset
	idx.ObjPos = ByteToUint64(buf[0:6])
	idx.ObjType = uint16(ByteToUint64(buf[6:8]))
	idx.ObjLen = ByteToUint64(buf[8:14])
	idx.ObjFlag = uint16(ByteToUint64(buf[14:16]))
	return idx
}

func (idx *Idx) Update(dst io.WriteSeeker) {
	buf := make([]byte, IdxSize)
	copy(buf[0:6], Uint64ToByte(idx.ObjPos)[2:])
	copy(buf[6:8], Uint64ToByte(uint64(idx.ObjType))[6:])
	copy(buf[8:14], Uint64ToByte(idx.ObjLen)[2:])
	copy(buf[14:16], Uint64ToByte(uint64(idx.ObjFlag))[6:])
	dst.Seek(idx.Offset, 0)
	dst.Write(buf)
}
