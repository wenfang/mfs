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
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		return
	}

	f, err := os.OpenFile(fname, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}

	var s mfs.Super
	s.MIdxSizeBit = 20
	s.SiteGrp = 0x00010000
	s.ImgLen = mfs.SuperSize + (1<<s.MIdxSizeBit)*mfs.IdxEntrySize
	s.ImgSize = uint64(fi.Size())
	s.NextId = mfs.FstMIdx * (1 << s.MIdxSizeBit)
  s.MIdx[0] = mfs.SuperSize
	s.Flush(f)
  fmt.Println(s.NextId)
  s.UpdateNextId(f)
  fmt.Println(s.NextId)
  s.UpdateNextId(f)
  fmt.Println(s.NextId)
  s.UpdateImgLen(f, 6234234)
  s.UpdateImgLen(f, 234455644)
}
