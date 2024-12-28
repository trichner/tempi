//go:build rp2040

package main

import (
	"fmt"
	"image/color"
	"machine"
	"strconv"
	"time"

	"github.com/trichner/tempi/pkg/adafruit4026"
	"github.com/trichner/tempi/pkg/adafruit4650"
	"github.com/trichner/tempi/pkg/hi"
	"github.com/trichner/tempi/pkg/logger"
	"github.com/trichner/tempi/pkg/pcf8523"
	"github.com/trichner/tempi/pkg/sht4x"
	"github.com/trichner/tempi/pkg/toggler"

	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"
)

var constWhite = color.RGBA{255, 255, 255, 0}

const watchDogMillis = 5000

const withSoilSensor = false

func main() {
	machine.InitSerial()

	log("setting up watchdog")
	wd := machine.Watchdog
	wd.Configure(machine.WatchdogConfig{watchDogMillis})

	time.Sleep(2 * time.Second)
	log("ready to go")

	log("setup i2c")
	bus := machine.I2C1
	err := bus.Configure(machine.I2CConfig{})
	if err != nil {
		panic(err)
	}

	log("setup RTC")
	rtc := pcf8523.New(bus, 0)

	log("setting power management")
	err = rtc.SetPowerManagement(pcf8523.PowerManagement_SwitchOver_ModeStandard_LowDetection)
	if err != nil {
		panic(err)
	}

	log("setup temp")
	sht := sht4x.New(bus, 0)

	var soilsensor adafruit4026.Device
	if withSoilSensor {
		log("setup soilsensor")
		soilsensor = adafruit4026.New(bus)
	}

	log("setup display")
	disp := adafruit4650.New(bus)

	log("configuring")
	err = disp.Configure()
	if err != nil {
		panic(err)
	}

	time.Sleep(100 * time.Millisecond)
	// err = disp.ClearDisplay()
	log("displaying")
	err = disp.Display()
	if err != nil {
		panic(err)
	}
	time.Sleep(100 * time.Millisecond)

	log("writing line")
	tinyfont.WriteLine(&disp, &freemono.Regular9pt7b, 0, 32, "Tempi", constWhite)
	err = disp.Display()
	if err != nil {
		panic(err)
	}

	log("setup SD card")
	lg, err := logger.New()
	if err != nil {
		tinyfont.WriteLine(&disp, &freemono.Regular9pt7b, 0, 15, "ERROR: setup SD card failed", constWhite)
		disp.Display()
		panic(err)
	}

	n, err := lg.IncrementBootCount()
	if err != nil {
		tinyfont.WriteLine(&disp, &freemono.Regular9pt7b, 0, 15, "ERROR: writing boot counter", constWhite)
		disp.Display()
		panic(err)
	}
	log("bootcount: " + strconv.Itoa(n))

	log("ready for blink")
	led := toggler.SetupToggler(machine.LED)

	lastMeasurement := time.Time{}

	buttons := []machine.Pin{machine.GPIO7, machine.GPIO8, machine.GPIO9}
	for _, b := range buttons {
		b.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	}
	buttonPressed := time.Time{}

	log("waiting a bit")
	time.Sleep(50 * time.Millisecond)

	log("starting watchdog")
	err = wd.Start()
	if err != nil {
		tinyfont.WriteLine(&disp, &freemono.Regular9pt7b, 0, 15, "ERROR: starting watchdog", constWhite)
		disp.Display()
		panic(err)
	}
	log("starting loop")

	for {
		wd.Update()
		led.Toggle()

		now, err := rtc.ReadTime()
		if err != nil {
			tinyfont.WriteLine(&disp, &freemono.Regular9pt7b, 0, 15, "ERROR: reading RTC", constWhite)
			disp.Display()
			panic(err)
		}

		var soilhum uint16
		if withSoilSensor {
			_, err = soilsensor.ReadMoisture()
			if err != nil {
				log("soil sensor failed to read: " + err.Error())
			}
			soilhum = soilsensor.AvgMoisture()
		}

		temp, hum, err := sht.ReadTemperatureHumidity()
		if err != nil {
			tinyfont.WriteLine(&disp, &freemono.Regular9pt7b, 0, 15, "ERROR: reading temp/hum", constWhite)
			disp.Display()
			panic(err)
		}

		if now.Sub(lastMeasurement) >= time.Minute*5 {
			log("appending record")
			lastMeasurement = now
			err = lg.AppendRecord(&logger.Record{
				Timestamp:                    now,
				MilliDegreeCelsius:           temp,
				MilliPercentRelativeHumidity: hum,
				SoilHumidity:                 int32(soilhum),
			})
			if err != nil {
				tinyfont.WriteLine(&disp, &freemono.Regular9pt7b, 0, 15, "ERROR: writing record", constWhite)
				disp.Display()
				panic(err)
			}
		}

		for _, b := range buttons {
			if !b.Get() {
				buttonPressed = now
			}
		}

		if now.Sub(buttonPressed) <= time.Second*40 {
			err = updateDisplay(&disp, now, temp, hum, soilhum)
			if err != nil {
				panic(err)
			}
		} else {
			err = disp.ClearDisplay()
			if err != nil {
				panic(err)
			}
		}

		time.Sleep(50 * time.Millisecond)
	}
}

func updateDisplay(disp *adafruit4650.Device, t time.Time, milliTemp, milliRh int32, soilHumidity uint16) error {
	hours := (t.Hour() + 2) % 24 // UTC -> CEST
	l := fmt.Sprintf("%02d:%02d:%02d", hours, t.Minute(), t.Second())

	deg := float32(milliTemp) / 1000.0
	lineTemp := fmt.Sprintf("%2.1f°C", deg)

	eff := hi.HeatIndexToEffect(hi.Calculate(milliTemp, milliRh))

	emoji := ""
	switch eff {
	case hi.EffectNone:
		fallthrough
	case hi.EffectUnknown:
		emoji = ":)"
	case hi.EffectCaution:
		emoji = ":/"
	case hi.EffectExtremeCaution:
		emoji = ":O"
	}

	rhum := float32(milliRh) / 1000.0
	lineRhum := fmt.Sprintf("%2.1f%%RH  %s", rhum, emoji)

	lineSoilHum := fmt.Sprintf("%d sh", soilHumidity)

	disp.ClearBuffer()
	tinyfont.WriteLine(disp, &freemono.Regular9pt7b, 0, 15, lineTemp, constWhite)
	tinyfont.WriteLine(disp, &freemono.Regular9pt7b, 0, 30, lineRhum, constWhite)
	tinyfont.WriteLine(disp, &freemono.Regular9pt7b, 0, 45, lineSoilHum, constWhite)
	tinyfont.WriteLine(disp, &freemono.Regular9pt7b, 0, 60, l, constWhite)
	return disp.Display()
}

func log(s string) {
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
