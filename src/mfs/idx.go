package mfs

import (
	"io"
)

const (
	IdxEntrySize = 16
)

type IdxEntry struct {
	IdxOff  uint64
	ObjPos  uint64
	ObjType uint16
	ObjLen  uint64
	ObjFlag uint16
}

func NewIdxEntry(f io.ReadSeeker, idxOff uint64) *IdxEntry {
	f.Seek(int64(idxOff), 0)
	buf := make([]byte, IdxEntrySize)
	if _, err := f.Read(buf); err != nil {
		return nil
	}

	entry := new(IdxEntry)
	entry.IdxOff = idxOff
	entry.ObjPos = ByteToUint64(buf[0:6])
	entry.ObjType = uint16(ByteToUint64(buf[6:8]))
	entry.ObjLen = ByteToUint64(buf[8:14])
	entry.ObjFlag = uint16(ByteToUint64(buf[14:16]))
	return entry
}

func (entry *IdxEntry) Update(f io.WriteSeeker) {
	buf := make([]byte, IdxEntrySize)
	copy(buf[0:6], Uint64ToByte(entry.ObjPos)[2:])
	copy(buf[6:8], Uint64ToByte(uint64(entry.ObjType))[6:])
	copy(buf[8:14], Uint64ToByte(entry.ObjLen)[2:])
	copy(buf[14:16], Uint64ToByte(uint64(entry.ObjFlag))[6:])
	f.Seek(int64(entry.IdxOff), 0)
	f.Write(buf)
}
