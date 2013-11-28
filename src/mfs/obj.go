package mfs

const (
	ObjHeadSize = 24
	ObjTailSize = 8
)

type ObjHead struct {
	ObjId   uint32
	ObjSize uint64
	ObjType uint16
	ObjLen  uint64
	objFlag uint16
}

type ObjTail struct {
	CRC32 uint32
}
