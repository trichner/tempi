package main

import (
	"fmt"
	"github.com/trichner/tempi/pkg/pcf8523"
	"machine"
	"time"
)

var now = time.Date(2023, 9, 16, 20, 34, 0, 0, time.UTC)

// main just set's a pcf8523 RTC to a hardcoded timestamp
func main() {

	machine.InitSerial()

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

	err = setPowerModeToSwitchOver(&rtc)
	if err != nil {
		panic(err)
	}

	if err := setRtc(&rtc, now); err != nil {
		panic(err)
	}

	panic(waitAndSetRtcOnButton(&rtc, now))
}

func waitAndSetRtcOnButton(rtc *pcf8523.Device, t time.Time) error {
	buttons := []machine.Pin{machine.GPIO7, machine.GPIO8, machine.GPIO9}
	for _, b := range buttons {
		b.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	}
	for {
		for i := range buttons {
			if buttons[i].Get() {
				if err := setRtc(rtc, t); err != nil {
					return err
				}
			}
		}
	}
}

func setPowerModeToSwitchOver(rtc *pcf8523.Device) error {
	log("setting power management")
	return rtc.SetPowerManagement(pcf8523.PowerManagement_SwitchOver_ModeStandard_LowDetection)
}

func setRtc(rtc *pcf8523.Device, t time.Time) error {

	err := rtc.SetTime(t)
	if err != nil {
		panic(err)
	}

	log("checking RTC")
	rtcNow, err := rtc.ReadTime()
	if err != nil {
		panic(err)
	}
	if rtcNow.Sub(t) < 0 {
		log("error: current RTC time is older than what we've just set it to!")
		return fmt.Errorf("error: current RTC time is older than what we've just set it to: %s < %s", rtcNow.Format(time.RFC3339), t.Format(time.RFC3339))
	}
	log("done!")
	return nil
}

func log(s string) {
	_, err := machine.Serial.Write([]byte(s + "\n\r"))
	if err != nil {
		panic(err)
	}
}
