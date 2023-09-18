package pcf8523

import (
	"testing"
)

func TestDecToBcd_RoundTrip(t *testing.T) {

	for i := 0; i < 60; i++ {
		a := bcd2bin(bin2bcd(i))
		if a != i {
			t.Logf("not equal: %d != %d", a, i)
			t.FailNow()
		}
	}
}
