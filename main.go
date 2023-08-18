package main

import (
	"github.com/trichner/tempi/pkg/adafruit4650"
	"image/color"
	"machine"
	"time"
)

const DISPLAY_ADDR = 0x3C

var constWhite = color.RGBA{255, 255, 255, 0}

func main() {

	bus := machine.I2C1
	err := bus.Configure(machine.I2CConfig{})
	if err != nil {
		panic(err)
	}

	disp := adafruit4650.NewI2C(bus)

	err = disp.Configure(adafruit4650.Config{
		Width:    128,
		Height:   64,
		VccState: adafruit4650.EXTERNALVCC,
	})
	if err != nil {
		panic(err)
	}

	time.Sleep(100 * time.Millisecond)
	err = disp.ClearDisplay()
	if err != nil {
		panic(err)
	}
	time.Sleep(100 * time.Millisecond)

	//for i := int16(0); i < 128; i++ {
	//	disp.SetPixel(i, 32, color.RGBA{R: 1})
	//}
	//for i := int16(0); i < 64; i++ {
	//	disp.SetPixel(64, i, color.RGBA{R: 1})
	//}
	err = disp.Display()
	if err != nil {
		panic(err)
	}

	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	for {
		led.Low()
		time.Sleep(time.Millisecond * 100)

		led.High()
		time.Sleep(time.Millisecond * 100)
	}
}
