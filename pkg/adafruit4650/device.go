// Package sh1106 implements a driver for the SH1106 display controller
//
// Copied from https://github.com/toyo/tinygo-sh1106 (under BSD 3-clause license)
package adafruit4650 // import "tinygo.org/x/drivers/sh1106"

import (
	"errors"
	"fmt"
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
func (d *Device) Configure(cfg Config) error {
	if cfg.Width != 0 {
		d.width = cfg.Width
	} else {
		d.width = 128
	}
	if cfg.Height != 0 {
		d.height = cfg.Height
	} else {
		d.height = 64
	}
	if cfg.VccState != 0 {
		d.vccState = cfg.VccState
	} else {
		d.vccState = SWITCHCAPVCC
	}

	bufferSize := d.width * d.height / 8
	d.buffer = make([]byte, bufferSize)

	time.Sleep(100 * time.Nanosecond)

	// https://github.com/adafruit/Adafruit_CircuitPython_DisplayIO_SH1107/blob/dad14cdeaaa38c8ca2b168c455cdc15f91d1cb0b/adafruit_displayio_sh1107.py#L113
	initSequence := []byte{
		0xae, // display off, sleep mode
		0xdc,
		0x00, // set display start line 0 (POR=0)
		0x81,
		0x4f, // contrast setting = 0x4f
		0x21, // vertical (column) addressing mode (POR=0x20)
		0xa0, // segment remap = 1 (POR=0, down rotation)
		0xc0, // common output scan direction = 0 (0 to n-1 (POR=0))
		0xa8,
		0x3f, // multiplex ratio = 128 (POR=0x3F) = height - 1
		0xd3,
		0x60, // set display offset mode = 0x60
		//0xd5 0x51  // divide ratio/oscillator: divide by 2, fOsc (POR)
		0xd9,
		0x22, // pre-charge/dis-charge period mode: 2 DCLKs/2 DCLKs (POR)
		0xdb,
		0x35, // VCOM deselect level = 0.770 (POR)
		//0xb0  // set page address = 0 (POR)
		0xa4, // entire display off, retain RAM, normal status (POR)
		0xa6, // normal (not reversed) display
	}

	for _, cmd := range initSequence {
		err := d.writeCommand(cmd)
		if err != nil {
			return err
		}
	}

	time.Sleep(100 * time.Millisecond)
	return d.writeCommand(
		0xaf, // DISPLAY_ON
	)
}

// ClearBuffer clears the image buffer
func (d *Device) ClearBuffer() {
	bzero(d.buffer)
}

// ClearDisplay clears the image buffer and clear the display
func (d *Device) ClearDisplay() error {
	d.ClearBuffer()
	return d.Display()
}

// SetPixel enables or disables a pixel in the buffer
// color.RGBA{0, 0, 0, x} is considered 'off', anything else
// with turn a pixel on the screen
func (d *Device) SetPixel(x int16, y int16, c color.RGBA) {
	if x < 0 || x >= d.width || y < 0 || y >= d.height {
		panic("out of range!")
		return
	}

	stride := d.width / 8
	byteIndex := x/8 + y*stride
	//   |   ----> x
	// y v
	//      p0           p1 .... p15
	//   0  a0 a1 .. a7  a0 a1 ..
	//   1  b0 b1 .. b7  b0 b1 ..
	//   2  c0 c1 .. c7
	//  ..
	//  64
	//
	if (c.R | c.G | c.B) != 0 {
		d.buffer[byteIndex] |= 1 << uint8(x%8)
	} else {
		d.buffer[byteIndex] &^= 1 << uint8(x%8)
	}
}

// Display sends the whole buffer to the screen
func (d *Device) Display() error {

	fmt.Printf("buffer len: %d x %d = %d\n", d.width, d.height, len(d.buffer))
	err := d.setPageAddress(uint8(0))
	if err != nil {
		return err
	}
	err = d.setColumnAddress(0)
	if err != nil {
		return err
	}

	//   |   ----> x
	// y v
	//      p0           p1 .... p15
	//   0  a0 a1 .. a7  a0 a1 ..
	//   1  b0 b1 .. b7  b0 b1 ..
	//   2  c0 c1 .. c7
	//  ..
	//  64
	//
	for column := int16(0); column < d.height; column++ {

		stride := d.width / 8
		offset := column * stride
		err = d.bus.WriteData(d.buffer[offset : offset+stride])
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Device) setPageAddress(p uint8) error {
	if p > 15 {
		panic("page out of bounds")
	}
	return d.writeCommand(0xB0 | (p & 0x07))
}

func (d *Device) setColumnAddress(column uint8) error {
	lo := column & 0b1111
	hi := (column >> 4) & 0b111
	err := d.writeCommand(SETLOWCOLUMN | lo)
	if err != nil {
		return err
	}
	return d.writeCommand(SETHIGHCOLUMN | hi)
}

// Size returns the current size of the display.
func (d *Device) Size() (w, h int16) {
	return d.width, d.height
}

// SetBuffer changes the whole buffer at once
func (d *Device) SetBuffer(buffer []byte) error {
	if len(buffer) != len(d.buffer) {
		return errors.New("wrong size buffer")
	}
	copy(d.buffer, buffer)
	return nil
}

func (d *Device) SetScroll(line int16) {
	d.writeCommand(SETSTARTLINE + uint8(line&0b111111))
}

// writeCommand sends a command to the display
func (d *Device) writeCommand(command uint8) error {
	return d.bus.WriteCommand(command)
}

func (d *Device) writeDoubleByteCommand(command, arg uint8) error {
	err := d.bus.WriteCommand(command)
	if err != nil {
		return err
	}
	return d.bus.WriteCommand(arg)
}

func bzero(buf []byte) {
	for i := range buf {
		buf[i] = 0
	}
}
