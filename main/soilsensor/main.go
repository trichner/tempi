//go:build rp2040

package main

import (
	"fmt"
	"machine"
	"time"

	"github.com/trichner/tempi/pkg/adafruit4026"
)

func main() {
	machine.InitSerial()

	bus := machine.I2C1
	err := bus.Configure(machine.I2CConfig{})
	if err != nil {
		panic(err)
	}

	dev := adafruit4026.New(bus)

	for {

		v, err := dev.ReadMoisture()
		if err != nil {
			panic(err)
		}
		fmt.Printf("%d moist\n", v)

		time.Sleep(time.Second)
	}
}
