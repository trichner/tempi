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
	img          draw.Image
	line         int
	bytesWritten int
}

func newMock() *mockBus {

	m := image.NewRGBA(image.Rect(0, 0, 128, 64))
	return &mockBus{img: m}
}

func (m *mockBus) WriteCommands(commands []byte) error {
	return nil
}

func (m *mockBus) WriteRAM(data []byte) error {

	m.bytesWritten += len(data)
	return nil

	//    *--> x
	//   y|    col0  col1  ... col127
	//    v p0  a0    b0         ..
	//          a1    b1         ..
	//          ..    ..         ..
	//          a7    b7         ..
	//      p1  a0    b0
	//          a1    b1
	//
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

	dev.Configure()

	for i := int16(0); i < 128; i++ {
		dev.SetPixel(i, 32, color.RGBA{R: 1})
	}
	for i := int16(0); i < 64; i++ {
		dev.SetPixel(64, i, color.RGBA{R: 1})
	}
	dev.Display()
	fmt.Println(bus.bytesWritten)
	bus.writeImage()
}
