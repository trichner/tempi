package main

import (
	"fmt"
	"github.com/trichner/tempi/pkg/colors"
	"io"
	"math"
	"os"
)

func main() {

	N := 256

	points := make([][]float64, N)
	for i := range points {
		x := float64(i)
		y1 := clamp(0, 255, fn(x))
		y2 := colors.Gamma(uint8(y1))
		points[i] = []float64{x, float64(y1), float64(y2)}
	}

	toFile(points)
	toGoFile(points)
}

func toGoFile(points [][]float64) error {
	f, err := os.Create("lut.go")
	if err != nil {
		return err
	}
	defer f.Close()

	discrete := make([]uint8, len(points))
	for i := range discrete {
		discrete[i] = uint8(points[i][1])
	}

	return writePointsGoCode(f, "main", "lut", discrete)
}
func toFile(points [][]float64) error {
	f, err := os.Create("points.dat")
	if err != nil {
		return err
	}
	defer f.Close()

	return writePoints(f, points)
}
func writePoints(w io.Writer, points [][]float64) error {
	for _, p := range points {
		first := true
		for _, y := range p {
			if first {
				first = false
			} else {
				fmt.Fprint(w, "\t")
			}
			_, err := fmt.Fprintf(w, "%f", y)
			if err != nil {
				return err
			}
		}
		fmt.Fprint(w, "\n")
	}
	return nil
}

func writePointsGoCode(w io.Writer, pkg, name string, points []uint8) error {
	fmt.Fprintf(w, "package %s\n\n", pkg)
	fmt.Fprintf(w, "var %s = [...]uint8{\n", name)
	/*(
	package main

	var lut = [...]uint8{
		0,
		3,
	 // ...
	}
	*/
	for _, p := range points {
		fmt.Fprintf(w, "  %d,\n", p)
	}
	fmt.Fprint(w, "}\n")
	return nil
}

func fn(x float64) float64 {

	yOffset := 20.0
	amplitude := 255.0 - yOffset
	xOffset := 0.5
	width := 0.18
	N := 255.0
	exp := -math.Pow(x/N-xOffset, 2) / (2 * math.Pow(width, 2))
	return amplitude*math.Pow(math.E, exp) + yOffset
}

func clamp(min, max, i float64) int {
	i = math.Round(i)
	return int(math.Max(min, math.Min(max, i)))
}
