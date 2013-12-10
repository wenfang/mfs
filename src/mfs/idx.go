package mfs

import (
	"errors"
	"io"
)

type Idx struct {
	Offset  int64
	ObjPos  uint64
	ObjType uint16
	ObjLen  uint64
	ObjFlag uint16
}

var (
	IdxErrRSeek = errors.New("Idx Read Seek Error")
	IdxErrRead  = errors.New("Idx Read Error")
	IdxErrWSeek = errors.New("Idx Write Seek Error")
	IdxErrWrite = errors.New("Idx Write Error")
)

// 从f中偏移量为Offset处读出Idx的内容
func NewIdx(f io.ReadSeeker, Offset int64) (*Idx, error) {
	buf := make([]byte, IdxSize)
	if _, err := f.Seek(Offset, 0); err != nil {
		return nil, IdxErrRSeek
	}
	if _, err := f.Read(buf); err != nil {
		return nil, IdxErrRead
	}

	idx := new(Idx)
	idx.Offset = Offset
	idx.ObjPos = ByteToUint64(buf[0:6])
	idx.ObjType = uint16(ByteToUint64(buf[6:8]))
	idx.ObjLen = ByteToUint64(buf[8:14])
	idx.ObjFlag = uint16(ByteToUint64(buf[14:16]))
	return idx, nil
}

// 将Idx内容写入f中，偏移量由idx.Offset指定
func (idx *Idx) Store(f io.WriteSeeker) error {
	buf := make([]byte, IdxSize)
	copy(buf[0:6], Uint64ToByte(idx.ObjPos)[2:])
	copy(buf[6:8], Uint64ToByte(uint64(idx.ObjType))[6:])
	copy(buf[8:14], Uint64ToByte(idx.ObjLen)[2:])
	copy(buf[14:16], Uint64ToByte(uint64(idx.ObjFlag))[6:])

	if _, err := f.Seek(idx.Offset, 0); err != nil {
		return IdxErrWSeek
	}
	if _, err := f.Write(buf); err != nil {
		return IdxErrWrite
	}
	return nil
}
