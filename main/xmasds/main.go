package main

import (
	"image/color"
	"machine"
	"time"

	"github.com/trichner/tempi/pkg/colors"
	"github.com/trichner/tempi/pkg/shims/rand"

	"tinygo.org/x/drivers/apa102"
)

var colorPalette = []color.RGBA{
	{0x7e, 0x12, 0x1d, 0xff},
	{0xbd, 0x36, 0x34, 0xff},
	{0xce, 0xac, 0x5c, 0xff},
	{0xe6, 0xdc, 0xb1, 0xff},
	{0x03, 0x4f, 0x1b, 0xff},
}

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
