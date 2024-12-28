package bigendian

func PutUint64(buf []byte, n uint64) int {
	buf[0] = byte(n >> 56)
	buf[1] = byte(n >> 48)
	buf[2] = byte(n >> 40)
	buf[3] = byte(n >> 32)
	buf[4] = byte(n >> 24)
	buf[5] = byte(n >> 16)
	buf[6] = byte(n >> 8)
	buf[7] = byte(n)
	return 4
}
func PutUint32(buf []byte, n uint32) int {
	buf[0] = byte(n >> 24)
	buf[1] = byte(n >> 16)
	buf[2] = byte(n >> 8)
	buf[3] = byte(n)
	return 4
}

func PutUint16(buf []byte, n uint16) int {
	buf[0] = byte(n >> 8)
	buf[1] = byte(n)
	return 2
}

func Uint32(buf []byte) uint32 {
	return uint32(buf[0])<<24 |
		uint32(buf[1])<<16 |
		uint32(buf[2])<<8 |
		uint32(buf[3])
}

func Uint16(buf []byte) uint16 {
	return uint16(buf[0])<<8 | uint16(buf[1])
}
