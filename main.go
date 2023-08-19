package main

import (
	"github.com/trichner/tempi/pkg/adafruit4650"
	"image/color"
	"machine"
	"time"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"
)

const DISPLAY_ADDR = 0x3C

var constWhite = color.RGBA{255, 255, 255, 0}

func main() {
	machine.InitSerial()

	time.Sleep(2 * time.Second)
	Log("ready to go")

	bus := machine.I2C1
	err := bus.Configure(machine.I2CConfig{})
	if err != nil {
		panic(err)
	}

	disp := adafruit4650.New(bus, 0)

	Log("configuring")
	err = disp.Configure()
	if err != nil {
		panic(err)
	}

	time.Sleep(100 * time.Millisecond)
	//err = disp.ClearDisplay()
	Log("displaying")
	err = disp.Display()
	if err != nil {
		panic(err)
	}
	time.Sleep(100 * time.Millisecond)

	Log("writing line")
	tinyfont.WriteLine(&disp, &freemono.Regular9pt7b, 0, 32, "Hello World!", constWhite)

	err = disp.Display()
	if err != nil {
		panic(err)
	}

	Log("ready for blink")

	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	for {
		led.Low()
		time.Sleep(time.Millisecond * 100)

		led.High()
		time.Sleep(time.Millisecond * 100)
	}
}

func Log(s string) {
	_, err := machine.Serial.Write([]byte(s + "\n\r"))
	if err != nil {
		panic(err)
	}
}
