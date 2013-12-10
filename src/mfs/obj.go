package mfs

import (
	"errors"
	"hash/crc32"
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
	OENewSeek       = errors.New("Obj Seek Error")
	OENewRead       = errors.New("Obj Read Error")
	OEMagic         = errors.New("Obj Magic Error")
	OERetriveSeek   = errors.New("Obj Retrive Seek Error")
	OERetrive       = errors.New("Obj Retrive Data Error")
	OEReadTail      = errors.New("Obj Read Tail Error")
	OEReadTailMagic = errors.New("Obj Read Tail Magic Error")
	OEReadTailCRC   = errors.New("Obj Read Tail CRC Error")
	OEStoreHSeek    = errors.New("Obj StoreHead Seek Error")
	OEStoreHWrite   = errors.New("Obj StoreHead Write Error")
	OEStoreDSeek    = errors.New("Obj StoreData Seek Error")
	OEStoreDCopy    = errors.New("Obj StoreData Copy Error")
	OEStoreDWrite   = errors.New("Obj StoreData Write Error")
)

// 从f的Offset位置读入数据，构造obj
func NewObj(f io.ReadSeeker, Offset int64) (*Obj, error) {
	if _, err := f.Seek(Offset, 0); err != nil {
		return nil, OENewSeek
	}

	buf := make([]byte, ObjHeadSize)
	if _, err := f.Read(buf); err != nil {
		return nil, OENewRead
	}

	if string(buf[0:4]) != "OSTA" {
		return nil, OEMagic
	}

	obj := new(Obj)
	obj.Offset = Offset
	obj.ObjId = uint32(ByteToUint64(buf[4:8]))
	obj.ObjSize = ByteToUint64(buf[8:14])
	obj.ObjType = uint16(ByteToUint64(buf[14:16]))
	obj.ObjLen = ByteToUint64(buf[16:22])
	obj.ObjFlag = uint16(ByteToUint64(buf[22:24]))
	return obj, nil
}

// 从f中获取对象内容到c，检查crc校验
func (obj *Obj) Retrive(f io.ReadSeeker, c io.Writer) error {
	h := crc32.NewIEEE()
	if _, err := f.Seek(obj.Offset+ObjHeadSize, 0); err != nil {
		return OERetriveSeek
	}
	if _, err := io.CopyN(io.MultiWriter(h, c), f, int64(obj.ObjLen)); err != nil {
		return OERetrive
	}

	buf := make([]byte, ObjTailSize)
	if _, err := f.Read(buf); err != nil {
		return OEReadTail
	}
	if string(buf[0:4]) != "OEND" {
		return OEReadTailMagic
	}
	if h.Sum32() != uint32(ByteToUint64(buf[4:8])) {
		return OEReadTailCRC
	}

	return nil
}

// 存储对象头部到f
func (obj *Obj) StoreHead(f io.WriteSeeker) error {
	buf := make([]byte, ObjHeadSize)
	n := copy(buf, []byte("OSTA"))
	n += copy(buf[n:], Uint64ToByte(uint64(obj.ObjId))[4:])
	n += copy(buf[n:], Uint64ToByte(obj.ObjSize)[2:])
	n += copy(buf[n:], Uint64ToByte(uint64(obj.ObjType))[6:])
	n += copy(buf[n:], Uint64ToByte(obj.ObjLen)[2:])
	copy(buf[n:], Uint64ToByte(uint64(obj.ObjFlag))[6:])

	if _, err := f.Seek(obj.Offset, 0); err != nil {
		return OEStoreHSeek
	}
	if _, err := f.Write(buf); err != nil {
		return OEStoreHWrite
	}
	return nil
}

// 从b(接收buf，可能在内存也可能是临时文件)写对象数据到f，最后写对象尾部到f
func (obj *Obj) StoreData(b io.Reader, f io.WriteSeeker) error {
	h := crc32.NewIEEE()
	if _, err := f.Seek(obj.Offset+ObjHeadSize, 0); err != nil {
		return OEStoreDSeek
	}
	if _, err := io.CopyN(io.MultiWriter(h, f), b, int64(obj.ObjLen)); err != nil {
		return OEStoreDCopy
	}
	obj.CRC32 = h.Sum32()

	buf := make([]byte, ObjTailSize)
	n := copy(buf, []byte("OEND"))
	copy(buf[n:], Uint64ToByte(uint64(obj.CRC32))[4:])
	if _, err := f.Write(buf); err != nil {
		return OEStoreDWrite
	}
	return nil
}
