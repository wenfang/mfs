package mfs

import (
	"os"
)

type IdxEntry struct {
	ObjPos  uint64
	ObjType uint16
	ObjLen  uint64
	ObjFlag uint16
}

func NewIdxEntry(f *os.File, objOff uint64) *IdxEntry {
	f.Seek(int64(objOff), 0)
	buf := make([]byte, IdxEntrySize)
	if _, err := f.Read(buf); err != nil {
		return nil
	}

	entry := new(IdxEntry)
	entry.ObjPos = ByteToUint64(buf[0:6])
	entry.ObjType = uint16(ByteToUint64(buf[6:8]))
	entry.ObjLen = ByteToUint64(buf[8:14])
	entry.ObjFlag = uint16(ByteToUint64(buf[14:16]))
	return entry
}
