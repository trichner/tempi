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
const readDelay = time.Millisecond * 3

const maxRetries = 5
const retryDelay = time.Millisecond

type Device struct {
	dev *seesaw.Device
	buf []uint16
	idx uint8
}

func New(i2c drivers.I2C) Device {
	ss := seesaw.New(i2c)
	ss.Address = DefaultAddress

	return Device{
		dev: ss,
		buf: make([]uint16, 16),
	}
}

func (d *Device) SetAddress(addr uint16) {
	d.dev.Address = addr
}

func (d *Device) ReadMoisture() (uint16, error) {
	var buf [2]byte

	//Arduino driver does retry here adding 1ms up to five times. Indeed, the sensor does not seem to be
	//very reliable.
	var err error
	for i := 0; i < maxRetries; i++ {
		err = d.dev.Read(seesaw.ModuleTouchBase, seesaw.FunctionTouchChannelOffset, buf[:], readDelay)
		if err == nil {
			v := uint16(buf[0])<<8 | uint16(buf[1])
			d.writeValue(v)
			return v, nil
		}
		time.Sleep(retryDelay)
	}

	return 0, err
}

func (d *Device) writeValue(v uint16) {
	if v == 0 {
		return
	}
	d.buf[d.idx] = v
	d.idx = (d.idx + 1) % uint8(len(d.buf))
}

func (d *Device) AvgMoisture() uint16 {
	// this 'should' be fine, moisture values are well below 2000
	var sum uint32
	var n uint32

	for _, v := range d.buf {
		if v > 0 {
			sum += uint32(v)
			n++
		}
	}

	if n == 0 {
		return 0
	}

	return uint16(sum / n)
}
