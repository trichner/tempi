package ustrconv

func Uint32toString(num uint32) string {
	if num == 0 {
		return "0"
	}

	lo := num % 100_000
	hi := num / 100_000

	var buf [10]byte
	itoaHalf(buf[:], hi)
	itoaHalf(buf[5:], lo)

	for i := 0; i < len(buf); i++ {
		if buf[i] != '0' {
			return string(buf[i:])
		}
	}
	panic("expected non '0' bytes")
}

func Uint16toString(num uint16) string {
	if num == 0 {
		return "0"
	}
	var buf [5]byte
	itoaHalf(buf[:], uint32(num))

	for i := 0; i < len(buf); i++ {
		if buf[i] != '0' {
			return string(buf[i:])
		}
	}
	panic("expected non '0' bytes")
}

// https://stackoverflow.com/questions/7890194/optimized-itoa-function

// 0 <= val <= 99999
func itoaHalf(str []byte, i uint32) {
	f1_10000 := uint32((1 << 28) / 10000)

	// 2^28 / 10000 is 26843.5456, but 26843.75 is sufficiently close.
	tmp := i*(f1_10000+1) - (i / 4)

	for i := 0; i < 5; i++ {
		digit := (byte)(tmp >> 28)
		str[i] = '0' + digit
		tmp = (tmp & 0x0fffffff) * 10
	}
}
