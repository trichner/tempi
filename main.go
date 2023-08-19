package main

import (
	"github.com/trichner/tempi/pkg/adafruit4650"
	"github.com/trichner/tempi/pkg/pcf8523"
	"image/color"
	"machine"
	"time"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"
)

var constWhite = color.RGBA{255, 255, 255, 0}

func main() {
	machine.InitSerial()

	time.Sleep(2 * time.Second)
	Log("ready to go")

	Log("setup i2c")
	bus := machine.I2C1
	err := bus.Configure(machine.I2CConfig{})
	if err != nil {
		panic(err)
	}

	Log("setup RTC")
	rtc := pcf8523.New(bus, 0)

	Log("setting power management")
	err = rtc.SetPowerManagement(pcf8523.PowerManagement_SwitchOver_ModeStandard_LowDetection)
	if err != nil {
		panic(err)
	}

	//Setting up current time:
	//now := time.Date(2023, 8, 19, 23, 28, 0, 0, time.UTC)
	//err = rtc.SetTime(now)
	//if err != nil {
	//	panic(err)
	//}

	//Log("dumping RTC")
	//data, err := rtc.Dump()
	//if err != nil {
	//	panic(err)
	//}
	//Log(FmtSliceToHex(data))

	Log("setup display")
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

		t, err := rtc.ReadTime()
		if err != nil {
			panic(err)
		}
		Log("rtc: " + t.Format(time.RFC3339))
		time.Sleep(time.Second)
	}
}

func Log(s string) {
	_, err := machine.Serial.Write([]byte(s + "\n\r"))
	if err != nil {
		panic(err)
	}
}

var mapping = [16]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'}

func FmtSliceToHex(s []byte) string {
	formatted := make([]byte, len(s)*2)
	for i := 0; i < len(s); i++ {
		formatted[i*2] = mapping[s[i]>>4]
		formatted[i*2+1] = mapping[s[i]&0xf]
	}
	return string(formatted)
}
