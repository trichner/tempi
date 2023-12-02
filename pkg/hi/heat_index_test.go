package hi

import (
	"fmt"
	"testing"
)

func TestCalculate(t *testing.T) {
	// for temp := 0; temp < 50; temp += 3 {
	temp := 30
	for rh := 0; rh <= 100; rh += 5 {

		index := Calculate(int32(temp*1000), int32(rh*1000))
		fmt.Printf("t: %d, rh: %d%% = hi: %d\n", temp, rh, index)
	}
	//}
}
