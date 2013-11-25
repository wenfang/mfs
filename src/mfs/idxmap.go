package mfs

const IdxMapMaxEntry = 4096

type IdxMap struct {
	IdxPos []uint64
	Num    uint16
}
