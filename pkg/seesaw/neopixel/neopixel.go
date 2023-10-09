package neopixel

import (
	"errors"
	"github.com/trichner/tempi/pkg/seesaw"
	"image/color"
	"strconv"
	"time"
)

// seesawWriteDelay the seesaw is quite timing sensitive and times out if not given enough time,
// this is an empirically determined delay that seems to have good results
const seesawWriteDelay = time.Millisecond * 50

const encodedColorLength = 3

type Seesaw interface {

	// Write writes an entire array into a given module and function
	Write(module seesaw.ModuleBaseAddress, function seesaw.FunctionAddress, buf []byte) error
}

type Device struct {
	seesaw          Seesaw
	ledCount        int
	pin             uint8
	lastOperationAt time.Time
}

func New(dev Seesaw, pin uint8, ledCount int) (*Device, error) {

	if !checkBufferLength(ledCount) {
		return nil, errors.New("invalid pixel count: " + strconv.Itoa(ledCount))
	}

	pixel := &Device{
		seesaw:   dev,
		ledCount: ledCount,
		pin:      pin,
	}

	time.Sleep(seesawWriteDelay)

	err := pixel.setupPin()
	if err != nil {
		return nil, errors.New("failed to update pixel pin " + strconv.Itoa(int(pin)) + ": " + err.Error())
	}

	time.Sleep(seesawWriteDelay)

	err = pixel.setupLedCount()
	if err != nil {
		return nil, errors.New("failed to update pixel count " + strconv.Itoa(ledCount) + ": " + err.Error())
	}

	time.Sleep(seesawWriteDelay)

	return pixel, nil
}

func (s *Device) setupLedCount() error {

	lenBytes := calculateBufferLength(s.ledCount)
	buf := []byte{byte(lenBytes >> 8), byte(lenBytes & 0xFF)}
	return s.seesaw.Write(seesaw.ModuleNeoPixelBase, seesaw.FunctionNeopixelBufLength, buf)
}

func calculateBufferLength(ledCount int) int {
	return ledCount * encodedColorLength
}

func (s *Device) setupPin() error {
	return s.seesaw.Write(seesaw.ModuleNeoPixelBase, seesaw.FunctionNeopixelPin, []byte{s.pin})
}

// WriteColorAtOffset updates the color for a single LED at the given offset
func (s *Device) WriteColorAtOffset(offset uint16, color color.RGBA) error {

	var buf [encodedColorLength]byte
	putGRB(buf[:], color)
	byteOffset := offset * encodedColorLength
	return s.writeBuffer(byteOffset, buf[:])
}

// WriteColors writes the given colors to the seesaws NeoPixel buffer
func (s *Device) WriteColors(buf []color.RGBA) error {

	if len(buf) > s.ledCount {
		return errors.New("buffer too big " + strconv.Itoa(len(buf)) + ">" + strconv.Itoa(s.ledCount*encodedColorLength))
	}

	tx := make([]byte, encodedColorLength*len(buf))
	pos := 0
	for _, c := range buf {
		w := tx[pos:]
		n := putGRB(w, c)
		pos += n
	}

	// the seesaw can at most deal with 30 bytes according to the datasheet, but
	// crashes after 29 bytes. So we only send 29 data bytes at a time
	const chunkSize = 29

	// write the data chunk-by-chunk
	for i := 0; i < len(tx); i += chunkSize {
		toSend := tx[i:min(i+chunkSize, len(tx))]
		err := s.writeBuffer(uint16(i), toSend)
		if err != nil {
			return errors.New("failed to write NeoPixel buffer offset " + strconv.Itoa(i) + ": " + err.Error())
		}
	}

	return nil
}

func (s *Device) writeBuffer(byteOffset uint16, buf []byte) error {
	tx := make([]byte, 2+len(buf))
	tx[0] = uint8(byteOffset >> 8)
	tx[1] = uint8(byteOffset)
	copy(tx[2:], buf)
	return s.seesaw.Write(seesaw.ModuleNeoPixelBase, seesaw.FunctionNeopixelBuf, tx)
}

func (s *Device) ShowPixels() error {

	// at most every 300us
	// https://github.com/adafruit/Adafruit_Seesaw/blob/8a2dc5e0645239cb34e23a4b62c456436b098ab3/seesaw_neopixel.cpp#L109
	s.waitSinceLastOperation(time.Microsecond * 300)

	return s.seesaw.Write(seesaw.ModuleNeoPixelBase, seesaw.FunctionNeopixelShow, nil)
}

func (s *Device) waitSinceLastOperation(d time.Duration) {
	diff := time.Since(s.lastOperationAt)
	for diff < d {
		time.Sleep(50 * time.Microsecond)
		diff = time.Since(s.lastOperationAt)
	}
}

// checkBufferLength checks whether the length is supported by seesaw. This depends on the pixel type.
// The seesaw has built in NeoPixel support for up to 170 RGB or 127 RGBW pixels. The
// output pin as well as the communication protocol frequency are configurable. Note:
// older firmware is limited to 63 pixels max.
func checkBufferLength(l int) bool {
	const maxRgbPixelCount = 170
	return l <= maxRgbPixelCount
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func putGRB(buf []byte, color color.RGBA) int {
	buf[0] = color.G
	buf[1] = color.R
	buf[2] = color.B
	return encodedColorLength
}
