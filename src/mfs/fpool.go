package mfs

import (
	"os"
)

const (
	FPoolSize = 32
)

type FPool struct {
	ImgName string
	Pool    chan *os.File
}

func NewFPool(ImgName string) *FPool {
	p := new(FPool)
	p.ImgName = ImgName
	p.Pool = make(chan *os.File, FPoolSize)
	return p
}

func (p *FPool) Alloc() *os.File {
	var res *os.File
	select {
	case res = <-p.Pool:
	default:
		res, _ = os.Open(p.ImgName)
	}
	return res
}

func (p *FPool) Free(f *os.File) {
	select {
	case p.Pool <- f:
	default:
		f.Close()
	}
}
