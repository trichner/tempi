package adafruit4650

import "tinygo.org/x/drivers"

type Bus interface {
	WriteCommand(command byte) error
	WriteData(data []byte) error
}

type i2cbus struct {
	dev  drivers.I2C
	addr uint8
}

func (i *i2cbus) WriteCommand(command byte) error {
	return i.dev.WriteRegister(i.addr, 0x00, []byte{command})
}

func (i *i2cbus) WriteData(data []byte) error {
	return i.dev.WriteRegister(i.addr, 0x40, data)
}
