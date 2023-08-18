package adafruit4650

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"testing"
	"time"
)

type mockBus struct {
	img  draw.Image
	line int
}

func newMock() *mockBus {

	m := image.NewRGBA(image.Rect(0, 0, 128, 64))
	return &mockBus{img: m}
}

func (m *mockBus) WriteCommand(command byte) error {
	return nil
}

func (m *mockBus) WriteData(data []byte) error {

	//   |   ----> x
	// y v
	//      p0           p1 .... p15
	//   0  a0 a1 .. a7  a0 a1 ..
	//   1  b0 b1 .. b7  b0 b1 ..
	//   2  c0 c1 .. c7
	//  ..
	//  64
	//
	//fmt.Println("received %d buffer", len(data))
	for x := 0; x < 128; x++ {
		byteOffset := x / 8
		b := data[byteOffset]
		if b&(1<<(x%8)) != 0 {
			m.img.Set(x, m.line, color.White)
		} else {
			m.img.Set(x, m.line, color.Black)
		}
	}
	m.line++
	return nil
}

func (m *mockBus) writeImage() {

	container := image.NewRGBA(m.img.Bounds().Inset(-1))
	draw.Draw(container, container.Bounds(), image.NewUniform(color.RGBA{0, 255, 0, 255}), image.Point{}, draw.Over)
	draw.Draw(container, m.img.Bounds(), m.img, image.Point{0, 0}, draw.Over)

	f, err := os.OpenFile(fmt.Sprintf("%d.png", time.Now().Unix()), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = png.Encode(f, container)
	if err != nil {
		panic(err)
	}
}

func TestDevice_Display(t *testing.T) {
	bus := newMock()
	dev := &Device{
		bus: bus,
	}

	dev.Configure(Config{
		Width:    128,
		Height:   64,
		VccState: EXTERNALVCC,
	})
	for i := int16(0); i < 128; i++ {
		dev.SetPixel(i, 32, color.RGBA{R: 1})
	}
	for i := int16(0); i < 64; i++ {
		dev.SetPixel(64, i, color.RGBA{R: 1})
	}
	dev.Display()
	bus.writeImage()
}
