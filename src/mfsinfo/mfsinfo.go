package main

import (
	"flag"
	"fmt"
	"mfs"
	"os"
)

func main() {
	var fname string
	flag.StringVar(&fname, "f", "", "<Img File Name>")
	flag.Parse()
	if fname == "" {
		fmt.Printf("Usage of %s:\r\n", os.Args[0])
		flag.PrintDefaults()
		return
	}

	f, err := os.Open(fname)
	if err != nil {
		fmt.Printf("[ERROR] File %s Open Error\r\n", fname)
		return
	}
	defer f.Close()

	buf := make([]byte, mfs.SuperSize+mfs.MIdxL1Size)
	if _, err := f.Read(buf); err != nil {
		fmt.Printf("[ERROR] File %s Read Error\r\n", fname)
	}

	fmt.Println("======")
	sitegrp := mfs.ByteToUint64(buf[4:8])
	imgLen := mfs.ByteToUint64(buf[8:14])
	imgSize := mfs.ByteToUint64(buf[14:20])
	nextObjId := mfs.ByteToUint64(buf[20:24])
	fmt.Printf("Site Group ID: %d(%X)\r\n", sitegrp, sitegrp)
	fmt.Printf("Img Len: %d(%X)\r\n", imgLen, imgLen)
	fmt.Printf("Img Size: %d(%X)\r\n", imgSize, imgSize)
	fmt.Printf("NextObjId: %d(%X)\r\n", nextObjId, nextObjId)

	fmt.Println("======")
	for i := 0; i <= int(nextObjId/mfs.MIdxL2Num); i++ {
		midx := mfs.ByteToUint64(buf[mfs.SuperSize+i*8 : mfs.SuperSize+i*8+8])
		fmt.Printf("MIdx %d: %d(%X)\r\n", i, midx, midx)
	}

	buf = make([]byte, mfs.MIdxL2Size)
	if _, err := f.Read(buf); err != nil {
		fmt.Printf("[ERROR] %s Read idx Error\r\n", fname)
		return
	}
	for i := 0; i < int(nextObjId); i++ {
		fmt.Println("-----------------------------------------")
		objPos := mfs.ByteToUint64(buf[i*mfs.IdxSize : i*mfs.IdxSize+6])
		objType := mfs.ByteToUint64(buf[i*mfs.IdxSize+6 : i*mfs.IdxSize+8])
		objLen := mfs.ByteToUint64(buf[i*mfs.IdxSize+8 : i*mfs.IdxSize+14])
		objFlag := mfs.ByteToUint64(buf[i*mfs.IdxSize+14 : i*mfs.IdxSize+16])
		fmt.Printf("ObjId: %d, ObjPos: %X, objType: %X, objLen: %X, objFlag: %X\r\n", i, objPos, objType, objLen, objFlag)

		head := make([]byte, mfs.ObjHeadSize)
		f.Seek(int64(objPos), 0)
		if _, err := f.Read(head); err != nil {
			fmt.Printf("[ERROR] Read Object Head Error\r\n")
			return
		}
		osta := string(head[0:4])
		objId := mfs.ByteToUint64(head[4:8])
		objSize := mfs.ByteToUint64(head[8:14])
		objType = mfs.ByteToUint64(head[14:16])
		objLen = mfs.ByteToUint64(head[16:22])
		objFlag = mfs.ByteToUint64(head[22:24])
		fmt.Printf("Start Magic: %s, ObjId: %d, objSize: %X, objType: %X, objLen: %X, objFlag: %X\r\n", osta, objId, objSize, objType, objLen, objFlag)

    tail := make([]byte, mfs.ObjTailSize)
    f.Seek(int64(objPos+mfs.ObjHeadSize+objLen), 0)
    if _, err := f.Read(tail); err != nil {
      fmt.Printf("[ERROR] Read Object Tail Error\r\n")
      return
    }
    oend := string(tail[0:4])
    crc := mfs.ByteToUint64(tail[4:8])
    fmt.Printf("End Magic: %s, CRC32: %X\r\n", oend, crc)
	}
}
