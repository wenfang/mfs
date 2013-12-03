package mfs

import (
  "errors"
  "os"
)

const (
	ImgOK = iota
	ImgFail
	ImgStop
)

const (
  MetaSize = 
	MaxImg = 64
)

type imgFile struct {
	ImgName string
	ImgStat uint16
}

type Meta struct {
	SiteId  uint32
	ImgNum  uint16
	ImgFile [MaxImg]imgFile
}

func NewMeta(metaFile string) (*Meta, error) {
  f, err := os.Open(metaFile)
  if err != nil {
    return nil, errors.New("Meta Open File Error")
  }
  defer f.Close()

	meta := new(Meta)
	return meta, nil
}
