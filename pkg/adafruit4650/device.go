// Package sh1106 implements a driver for the SH1106 display controller
//
// Copied from https://github.com/toyo/tinygo-sh1106 (under BSD 3-clause license)
package adafruit4650 // import "tinygo.org/x/drivers/sh1106"

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers"
)

const DISPLAY_OFFSET_ADAFRUIT_FEATHERWING_OLED_4650 = 0x60

// Device wraps an SPI connection.
type Device struct {
	bus    Bus
	buffer []byte
	width  int16
	height int16
	//bufferSize int16
	vccState VccMode
}

// Config is the configuration for the display
type Config struct {
	Width    int16
	Height   int16
	VccState VccMode
}

type I2CBus struct {
	wire    drivers.I2C
	Address uint16
}

type SPIBus struct {
	wire     drivers.SPI
	dcPin    machine.Pin
	resetPin machine.Pin
	csPin    machine.Pin
}

type Buser interface {
	configure()
	tx(data []byte, isCommand bool)
	setAddress(address uint16)
}

type VccMode uint8

// NewI2C creates a new SSD1306 connection. The I2C wire must already be configured.
func NewI2C(bus drivers.I2C) Device {
	return Device{
		bus: &i2cbus{
			dev:  bus,
			addr: Address,
		},
	}
}

// Configure initializes the display with default configuration
func (d *Device) Configure() error {
	d.width = 128
	d.height = 64

	bufferSize := d.width * d.height / 8
	d.buffer = make([]byte, bufferSize)

	time.Sleep(100 * time.Millisecond)

	// https://github.com/adafruit/Adafruit_CircuitPython_DisplayIO_SH1107/blob/dad14cdeaaa38c8ca2b168c455cdc15f91d1cb0b/adafruit_displayio_sh1107.py#L113
	initSequence := []byte{
		0xae, // display off, sleep mode
		//0xd5, 0x41, // set display clock divider (from original datasheet)
		//0xd5, 0x51, // set display clock divider
		0xd5, 0x80, // set display clock divider
		0xd9, 0x22, // pre-charge/dis-charge period mode: 2 DCLKs/2 DCLKs (POR)
		0x20,       // memory mode
		0x81, 0x4f, // contrast setting = 0x4f
		0xad, 0x8a, // set dc/dc pump
		0xa0,       // segment remap, flip-x
		0xc0,       // common output scan direction
		0xdc, 0x00, // set display start line 0 (POR=0)
		0xa8, 0x3f, // multiplex ratio, height - 1 = 0x3f
		0xd3, 0x60, // set display offset mode = 0x60
		0xdb, 0x35, // VCOM deselect level = 0.770 (POR)
		0xa4, // entire display off, retain RAM, normal status (POR)
		0xa6, // normal (not reversed) display
		0xaf, // DISPLAY_ON
	}

	err := d.bus.WriteCommands(initSequence)
	if err != nil {
		return err
	}

	time.Sleep(100 * time.Millisecond)
	return nil
}

// ClearDisplay clears the image buffer and clear the display
func (d *Device) ClearDisplay() error {
	d.clearBuffer()
	return d.Display()
}

// clearBuffer clears the image buffer
func (d *Device) clearBuffer() {
	bzero(d.buffer)
}

// SetPixel enables or disables a pixel in the buffer
// color.RGBA{0, 0, 0, x} is considered 'off', anything else
// with turn a pixel on the screen
func (d *Device) SetPixel(x int16, y int16, c color.RGBA) {
	if x < 0 || x >= d.width || y < 0 || y >= d.height {
		return
	}

	//flip y
	y = d.height - y - 1

	page := x / 8
	bytesPerPage := d.height
	byteIndex := y + bytesPerPage*page
	bit := x % 8
	//    *-----> y
	//    |
	//   x|     col0  col1  ... col63
	//    v  p0  a0    b0         ..
	//           a1    b1         ..
	//           ..    ..         ..
	//           a7    b7         ..
	//       p1  a0    b0
	//           a1    b1
	//
	if (c.R | c.G | c.B) != 0 {
		d.buffer[byteIndex] |= 1 << uint8(bit)
	} else {
		d.buffer[byteIndex] &^= 1 << uint8(bit)
	}
}

// Display sends the whole buffer to the screen
func (d *Device) Display() error {

	bytesPerPage := d.height

	pages := (d.width + 7) / 8
	for page := int16(0); page < pages; page++ {

		err := d.setRAMPosition(uint8(page), 0)
		if err != nil {
			return err
		}

		offset := page * bytesPerPage
		err = d.bus.WriteRAM(d.buffer[offset : offset+bytesPerPage])
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Device) setRAMPosition(page uint8, column uint8) error {
	if page > 15 {
		panic("page out of bounds")
	}
	if column > 127 {
		panic("column out of bounds")
	}
	setPage := 0xB0 | (page & 0xF)

	lo := column & 0xF
	setLowColumn := SETLOWCOLUMN | lo

	hi := (column >> 4) & 0x7
	setHighColumn := SETHIGHCOLUMN | hi

	cmds := []byte{
		setPage,
		setLowColumn,
		setHighColumn,
	}

	return d.bus.WriteCommands(cmds)
}

// Size returns the current size of the display.
func (d *Device) Size() (w, h int16) {
	return d.width, d.height
}

func bzero(buf []byte) {
	for i := range buf {
		buf[i] = 0
	}
}
