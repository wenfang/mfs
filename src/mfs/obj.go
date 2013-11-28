package mfs

import (
	"io"
)

const (
	ObjHeadSize = 24
	ObjTailSize = 8
)

type Obj struct {
	Offset  int64
	ObjId   uint32
	ObjSize uint64
	ObjType uint16
	ObjLen  uint64
	ObjFlag uint16
	CRC32   uint32
}

func NewObj(f io.ReadSeeker, entry *IdxEntry) *Obj {
	o := new(Obj)
	return o
}

func (o *Obj) Store(src io.Reader, dst io.WriteSeeker) {
	buf := make([]byte, ObjHeadSize)
	n := copy(buf, []byte("OSTA"))
	n += copy(buf[n:], Uint64ToByte(uint64(o.ObjId))[4:])
	n += copy(buf[n:], Uint64ToByte(o.ObjSize)[4:])
	n += copy(buf[n:], Uint64ToByte(uint64(o.ObjType))[6:])
	n += copy(buf[n:], Uint64ToByte(o.ObjLen)[4:])
	n += copy(buf[n:], Uint64ToByte(uint64(o.ObjFlag))[6:])

	dst.Seek(o.Offset, 0)
	dst.Write(buf)

	io.CopyN(dst, src, int64(o.ObjLen))
}
