package ustrconv

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

func TestPutUint16(t *testing.T) {
	for i := 0; i <= math.MaxUint16; i++ {
		t.Run(fmt.Sprintf("strconv %d", i), func(t *testing.T) {
			expected := fmt.Sprintf("%d", i)
			actual := Uint16toString(uint16(i))
			if expected != actual {
				t.Fatalf("expected %s is not actual %s", expected, actual)
			}
		})
	}
}

func TestPutUint32(t *testing.T) {
	rnd := rand.New(rand.NewSource(1337))
	for i := 0; i <= 100_000; i++ {
		number := rnd.Uint32()
		t.Run(fmt.Sprintf("strconv %d", number), func(t *testing.T) {
			actual := Uint32toString(number)
			expected := fmt.Sprintf("%d", number)
			if expected != actual {
				t.Fatalf("expected %s is not actual %s", expected, actual)
			}
		})

	}
}

func BenchmarkPutUint16_Naive(b *testing.B) {
	rnd := rand.New(rand.NewSource(int64(b.N)))
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		number := uint16(rnd.Uint32())
		naiveUInt16ToString(number)
	}
}

func BenchmarkPutUint16_Terje(b *testing.B) {
	rnd := rand.New(rand.NewSource(int64(b.N)))
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		number := uint16(rnd.Uint32())
		Uint16toString(number)
	}
}

func naiveUInt16ToString(i uint16) string {
	if i == 0 {
		return "0"
	}

	result := make([]byte, 0, 10)
	for i > 0 {
		result = append([]byte{byte(i%10 + 48)}, result...)
		i /= 10
	}

	return string(result)
}
