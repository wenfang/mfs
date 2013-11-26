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

func NewIdxEntry(buf []byte) *IdxEntry {
	if len(buf) != IdxEntrySize {
		return nil
	}

	entry := new(IdxEntry)
	entry.ObjPos = byteTouint64(buf[0:6])
	entry.ObjType = uint16(byteTouint64(buf[6:8]))
	entry.ObjLen = byteTouint64(buf[8:14])
	entry.ObjFlag = uint16(byteTouint64(buf[14:16]))
	return entry
}

func NewIdxEntryFromFile(f *os.File) *IdxEntry {
	buf := make([]byte, IdxEntrySize)
	if _, err := f.Read(buf); err != nil {
		return nil
	}
	return NewIdxEntry(buf)
}
