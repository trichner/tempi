package main

import (
	"image/color"
	"machine"
	"time"

	"github.com/trichner/tempi/pkg/colors"
	"github.com/trichner/tempi/pkg/shims/rand"
	"github.com/trichner/tempi/pkg/ustrconv"

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

var colorPaletteCyberpunk = []color.RGBA{
	{0xff, 0x00, 0x7f, 0xff}, // Hot pink
	{0x00, 0xff, 0xff, 0xff}, // Cyan
	{0x9d, 0x00, 0xff, 0xff}, // Electric purple
	{0x00, 0xd4, 0xff, 0xff}, // Neon blue
	{0xff, 0x00, 0xff, 0xff}, // Magenta
}

var colorPaletteSynthwave = []color.RGBA{
	{0xff, 0x6a, 0xc1, 0xff}, // Pink
	{0x79, 0x3a, 0x80, 0xff}, // Deep purple
	{0x2d, 0x00, 0x4b, 0xff}, // Dark violet
	{0xf9, 0xc8, 0x0e, 0xff}, // Sun yellow
	{0xff, 0x35, 0x6b, 0xff}, // Sunset coral
}

var colorPaletteRubiks = []color.RGBA{
	{0xff, 0x00, 0x00, 0xff}, // Red
	{0x00, 0x9b, 0x48, 0xff}, // Green
	{0x00, 0x46, 0xad, 0xff}, // Blue
	{0xff, 0xd5, 0x00, 0xff}, // Yellow
	{0xff, 0x58, 0x00, 0xff}, // Orange
	{0xff, 0xff, 0xff, 0xff}, // White
}

var palettes = [][]color.RGBA{
	colorPaletteXmas,
	colorPaletteCherryBlossom,
	colorPaletteWinterWarmer,
	colorPaletteCyberpunk,
	colorPaletteSynthwave,
	colorPaletteRubiks,
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

	log("initializing ADC")
	adc := newADC(len(palettes))

	log("initializing apa102 driver")
	strip := apa102.New(spi)
	buf := make([]color.RGBA, 30*5)

	log("initializing leds")
	r := rand.New(rand.NewSource(1337))
	leds := make([]Led, 30*5)

	log("starting animation")

	for {
		if s, changed := adc.GetSector(); changed {
			log("palette changed to " + ustrconv.Uint16toString(uint16(s)))
			setLeds(leds, r, palettes[s])
		}

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

func setLeds(leds []Led, r *rand.Rand, palette []color.RGBA) {
	for i := range leds {
		leds[i].Brightness = uint8(r.Uint64())
		leds[i].Color = palette[r.Intn(len(palette))]
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
