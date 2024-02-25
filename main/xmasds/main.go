package main

import (
	"image/color"
	"machine"
	"time"

	"github.com/trichner/tempi/pkg/colors"
	"github.com/trichner/tempi/pkg/shims/rand"

	"tinygo.org/x/drivers/apa102"
)

var colorPaletteXmas = []color.RGBA{
	{0x7e, 0x12, 0x1d, 0xff},
	{0xbd, 0x36, 0x34, 0xff},
	{0xce, 0xac, 0x5c, 0xff},
	{0xe6, 0xdc, 0xb1, 0xff},
	{0x03, 0x4f, 0x1b, 0xff},
}

var colorPaletteCherryBlossom = []color.RGBA{
	{0xec, 0x27, 0x5f, 0xff},
	{0xf2, 0x54, 0x77, 0xff},
	{0xff, 0xa7, 0xa6, 0xff},
	{0xff, 0xdc, 0xdc, 0xff},
	{0xd4, 0xe0, 0xee, 0xff},
}

var colorPaletteWinterWarmer = []color.RGBA{
	{0x55, 0x18, 0x25, 0xff},
	{0xa0, 0x44, 0x3f, 0xff},
	{0xad, 0x70, 0x6c, 0xff},
	{0xed, 0x7d, 0x4b, 0xff},
	{0xe9, 0xc6, 0x8a, 0xff},
	//{0xf4, 0xde, 0xb9, 0xff},
	//{0xf2, 0xf0, 0xf0, 0xff},
}

var colorPalette = colorPaletteWinterWarmer

func main() {
	machine.InitSerial()
	time.Sleep(2 * time.Second)

	log("booted")

	spi := machine.SPI0
	err := spi.Configure(machine.SPIConfig{Frequency: 500_000})
	if err != nil {
		panic(err)
	}

	log("spi configured")

	strip := apa102.New(spi)
	leds := make([]Led, 30*5)
	buf := make([]color.RGBA, 30*5)

	r := rand.New(rand.NewSource(1337))

	log("initializing leds")

	for i := range leds {
		leds[i].Brightness = uint8(r.Uint64())
		leds[i].Color = colorPalette[r.Intn(len(colorPalette))]
	}

	log("starting animation")

	for {
		for i := range leds {
			buf[i] = leds[i].Next()
		}

		_, err := strip.WriteColors(buf)
		if err != nil {
			panic(err)
		}

		time.Sleep(10 * time.Millisecond)
	}
}

func log(s string) {
	_, err := machine.Serial.Write([]byte(s + "\n\r"))
	if err != nil {
		panic(err)
	}
}

type Led struct {
	Color      color.RGBA
	Brightness uint8
}

func (l *Led) Next() color.RGBA {
	b := colors.Sin8i(l.Brightness)
	l.Color.A = b
	// l.Color.A = l.Brightness
	l.Brightness++
	return colors.GammaCorrect(l.Color)
}
