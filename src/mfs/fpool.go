package mfs

import (
	"os"
)

const (
	fPoolSize = 32
)

type FPool interface {
  Alloc() *os.File
  Free(*os.File)
}

type fPool struct {
	ImgName string
	Pool    chan *os.File
}

func NewFPool(ImgName string) FPool {
	return &fPool{ImgName, make(chan *os.File, fPoolSize)}
}

func (p *fPool) Alloc() *os.File {
	var res *os.File
	select {
	case res = <-p.Pool:
	default:
		res, _ = os.Open(p.ImgName)
	}
	return res
}

func (p *fPool) Free(f *os.File) {
	select {
	case p.Pool <- f:
	default:
		f.Close()
	}
}
