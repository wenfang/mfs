package mfs

import (
    "os"
)

type FPool struct {
    ImgName string
    Pool    chan *os.File
}
