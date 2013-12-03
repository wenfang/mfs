package main

import (
	"flag"
	"fmt"
	"mfs"
	"os"
)

func main() {
	var fname string
	var sitegrp uint64
	flag.StringVar(&fname, "f", "", "<Img FileName>")
	flag.Uint64Var(&sitegrp, "s", 0, "<SiteGroup Id>")
	flag.Parse()
	if fname == "" || sitegrp == 0 {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		return
	}

	fw, err := os.OpenFile(fname, os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("[ERROR] Open %s Failed\n", fname)
		return
	}
	defer fw.Close()

	fi, err := fw.Stat()
	if err != nil {
		fmt.Printf("[ERROR] Stat %s Failed\n", fname)
		return
	}

	buf := make([]byte, mfs.SuperBlockSize+mfs.MIdxBlockSize)
	n := copy(buf, []byte("MJFS"))
	n += copy(buf[n:], mfs.Uint64ToByte(sitegrp)[4:])
	n += copy(buf[n:], mfs.Uint64ToByte(mfs.SuperBlockSize + mfs.MIdxBlockSize + mfs.MIdxSize*mfs.IdxSize)[2:])
	n += copy(buf[n:], mfs.Uint64ToByte(uint64(fi.Size()))[2:])
	copy(buf[n:], mfs.Uint64ToByte(0)[4:])

	copy(buf[mfs.SuperBlockSize:], mfs.Uint64ToByte(mfs.SuperBlockSize+mfs.MIdxBlockSize))

	fw.Seek(0, 0)
	fw.Write(buf)
}
