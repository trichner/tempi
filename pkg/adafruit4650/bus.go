package adafruit4650

import "tinygo.org/x/drivers"

type Bus interface {
	WriteCommands(commands []byte) error
	WriteRAM(data []byte) error
}

type i2cbus struct {
	dev  drivers.I2C
	addr uint8
}

func (i *i2cbus) WriteCommands(commands []byte) error {
	// 0x00 is the control byte for 'only command data following'
	return i.dev.WriteRegister(i.addr, 0x00, commands)
}

func (i *i2cbus) WriteRAM(data []byte) error {
	// 0x40 is the control byte for 'only RAM data following'
	return i.dev.WriteRegister(i.addr, 0x40, data)
}
