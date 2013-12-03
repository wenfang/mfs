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

  buf := make([]byte, mfs.SuperBlockSize + mfs.MIdxBlockSize)
  if _, err := f.Read(buf); err != nil {
    fmt.Printf("[ERROR] File %s Read Error\r\n", fname)
  }

  fmt.Printf("Site Group ID: %d(%X)\r\n", mfs.ByteToUint64(buf[4:8]), mfs.ByteToUint64(buf[4:8]))
  fmt.Printf("Img Len: %d(%X)\r\n", mfs.ByteToUint64(buf[8:14]), mfs.ByteToUint64(buf[8:14]))
  fmt.Printf("Img Size: %d(%X)\r\n", mfs.ByteToUint64(buf[14:20]), mfs.ByteToUint64(buf[14:20]))
  fmt.Printf("NextObjId: %d(%X)\r\n", mfs.ByteToUint64(buf[20:24]), mfs.ByteToUint64(buf[20:24]))
}
