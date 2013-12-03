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

	buf := make([]byte, mfs.SuperBlockSize+mfs.MIdxBlockSize)
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
	for i := 0; i <= int(nextObjId/mfs.MIdxSize); i++ {
		midx := mfs.ByteToUint64(buf[mfs.SuperBlockSize+i*8 : mfs.SuperBlockSize+i*8+8])
		fmt.Printf("MIdx %d: %d(%X)\r\n", i, midx, midx)
	}

	fmt.Println("======")

	buf = make([]byte, mfs.MIdxSize*mfs.IdxSize)
	if _, err := f.Read(buf); err != nil {
		fmt.Printf("[ERROR] %s Read idx Error\r\n", fname)
	}
	for i := 0; i < int(nextObjId); i++ {
		objPos := mfs.ByteToUint64(buf[i*mfs.IdxSize : i*mfs.IdxSize+6])
		objType := mfs.ByteToUint64(buf[i*mfs.IdxSize+6 : i*mfs.IdxSize+8])
		objLen := mfs.ByteToUint64(buf[i*mfs.IdxSize+8 : i*mfs.IdxSize+14])
		objFlag := mfs.ByteToUint64(buf[i*mfs.IdxSize+14 : i*mfs.IdxSize+16])
    fmt.Printf("ObjId: %d, ObjPos: %X, objType: %X, objLen: %X, objFlag: %X\r\n", i, objPos, objType, objLen, objFlag)
	}
}
