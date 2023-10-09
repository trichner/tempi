// Package adafruit4026 implements a driver for the Adafruit 4026 capacitive moisture sensor
//
// Datasheet: https://cdn-learn.adafruit.com/downloads/pdf/adafruit-stemma-soil-sensor-i2c-capacitive-moisture-sensor.pdf
package adafruit4026

import (
	"github.com/trichner/tempi/pkg/seesaw"
	"time"
	"tinygo.org/x/drivers"
)

const DefaultAddress = 0x36

// from the Arduino driver https://github.com/adafruit/Adafruit_Seesaw/blob/c3e7b8f4dfdcc1f8ca3c0cabbacfd441ba8f8212/Adafruit_seesaw.cpp#L362
const readDelay = time.Microsecond * 3000

type Device struct {
	dev *seesaw.Device
}

func New(i2c drivers.I2C) Device {
	ss := seesaw.New(i2c)
	ss.Address = DefaultAddress

	return Device{
		dev: ss,
	}
}

func (d *Device) SetAddress(addr uint16) {
	d.dev.Address = addr
}

func (d *Device) ReadMoisture() (uint16, error) {
	var buf [2]byte

	//TODO: Arduino driver does retries here adding 1ms up to five times
	err := d.dev.Read(seesaw.ModuleTouchBase, seesaw.FunctionTouchChannelOffset, buf[:], readDelay)
	if err != nil {
		return 0, err
	}
	v := uint16(buf[0])<<8 | uint16(buf[1])
	return v, nil
}
