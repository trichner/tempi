package main

import (
	"fmt"
	"github.com/trichner/tempi/pkg/adafruit4650"
	"github.com/trichner/tempi/pkg/logger"
	"github.com/trichner/tempi/pkg/pcf8523"
	"github.com/trichner/tempi/pkg/sht4x"
	"image/color"
	"machine"
	"strconv"
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
	//now := time.Date(2023, 8, 24, 20, 49, 0, 0, time.UTC)
	//err = rtc.SetTime(now)
	//if err != nil {
	//	panic(err)
	//}

	//Log("dumping RTC")
	//data, err := rtc.Dump()
	//if err != nil {
	//	panic(err)
	//}

	Log("setup temp")
	sht := sht4x.New(bus, 0)

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

	Log("setup SD card")
	lg, err := logger.New()
	if err != nil {
		panic(err)
	}

	n, err := lg.IncrementBootCount()
	if err != nil {
		panic(err)
	}
	Log("bootcount: " + strconv.Itoa(n))

	Log("ready for blink")

	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	toggle := 0
	lastMeasurement := time.Time{}
	for {
		toggle++
		if toggle%2 == 0 {
			led.Low()
		} else {
			led.High()
		}

		t, err := rtc.ReadTime()
		if err != nil {
			panic(err)
		}

		temp, hum, err := sht.ReadTemperatureHumidity()
		if err != nil {
			panic(err)
		}

		if t.Sub(lastMeasurement) >= time.Minute*5 {
			Log("appending record")
			lastMeasurement = t
			err = lg.AppendRecord(&logger.Record{
				Timestamp:                    t,
				MilliDegreeCelsius:           temp,
				MilliPercentRelativeHumidity: hum,
			})
			if err != nil {
				panic(err)
			}
		}

		err = updateDisplay(&disp, t, temp, hum)
		if err != nil {
			panic(err)
		}

		time.Sleep(50 * time.Millisecond)
	}
}

func updateDisplay(disp *adafruit4650.Device, t time.Time, milliTemp, milliRh int32) error {

	hours := (t.Hour() + 2) % 24 // UTC -> CEST
	l := fmt.Sprintf("%02d:%02d:%02d", hours, t.Minute(), t.Second())

	deg := float32(milliTemp) / 1000.0
	lineTemp := fmt.Sprintf("%2.1f°C", deg)

	rhum := float32(milliRh) / 1000.0
	lineRhum := fmt.Sprintf("%2.1f%%RH", rhum)

	disp.ClearBuffer()
	tinyfont.WriteLine(disp, &freemono.Regular9pt7b, 0, 50, l, constWhite)
	tinyfont.WriteLine(disp, &freemono.Regular9pt7b, 0, 10, lineTemp, constWhite)
	tinyfont.WriteLine(disp, &freemono.Regular9pt7b, 0, 25, lineRhum, constWhite)
	return disp.Display()
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
