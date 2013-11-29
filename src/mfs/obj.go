package mfs

import (
	"errors"
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

var (
	ObjErrRetrive = errors.New("Retrive Obj Data Error")
)

func NewObj(src io.ReadSeeker, Offset int64) *Obj {
	src.Seek(Offset, 0)
	buf := make([]byte, ObjHeadSize)
	if _, err := src.Read(buf); err != nil {
		return nil
	}

	if string(buf[0:4]) != "OSTA" {
		return nil
	}

	obj := new(Obj)
	obj.Offset = Offset
	obj.ObjId = uint32(ByteToUint64(buf[4:8]))
	obj.ObjSize = ByteToUint64(buf[8:14])
	obj.ObjType = uint16(ByteToUint64(buf[14:16]))
	obj.ObjLen = ByteToUint64(buf[16:22])
	obj.ObjFlag = uint16(ByteToUint64(buf[22:24]))
	return obj
}

// 从src中获取对象内容到dst
func (o *Obj) Retrive(src io.ReadSeeker, dst io.Writer) error {
	src.Seek(o.Offset+ObjHeadSize, 0)
	if _, err := io.CopyN(dst, src, int64(o.ObjLen)); err != nil {
		return ObjErrRetrive
	}

	return nil
}

func (o *Obj) StoreHead(dst io.WriteSeeker) {
	buf := make([]byte, ObjHeadSize)
	n := copy(buf, []byte("OSTA"))
	n += copy(buf[n:], Uint64ToByte(uint64(o.ObjId))[4:])
	n += copy(buf[n:], Uint64ToByte(o.ObjSize)[2:])
	n += copy(buf[n:], Uint64ToByte(uint64(o.ObjType))[6:])
	n += copy(buf[n:], Uint64ToByte(o.ObjLen)[2:])
	n += copy(buf[n:], Uint64ToByte(uint64(o.ObjFlag))[6:])
	dst.Seek(o.Offset, 0)
	dst.Write(buf)
}

func (o *Obj) StoreData(src io.Reader, dst io.WriteSeeker) {
	dst.Seek(o.Offset+ObjHeadSize, 0)
	io.CopyN(dst, src, int64(o.ObjLen))

	buf := make([]byte, ObjTailSize)
	n := copy(buf, []byte("OEND"))
	n += copy(buf[n:], Uint64ToByte(uint64(o.CRC32))[4:])
	dst.Write(buf)
}
