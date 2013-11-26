package mfs

func byteTouint64(buf []byte) uint64 {
	var res uint64
	for i, v := range buf {
		if i >= 8 {
			break
		}
		res <<= 8
		res += uint64(v)
	}
	return res
}

func uint64Tobyte(v uint64) []byte {
	buf := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		buf[i] = byte(v)
		if v >>= 8; v == 0 {
			break
		}
	}
	return buf
}
