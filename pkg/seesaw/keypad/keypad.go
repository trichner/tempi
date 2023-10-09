package keypad

import (
	"github.com/trichner/tempi/pkg/seesaw"
	"time"
	"unsafe"
)

type Seesaw interface {

	// Read reads a number of bytes from the device after sending the read command and waiting 'delay'. The delays depend
	// on the module and function and are documented in the seesaw datasheet
	Read(module seesaw.ModuleBaseAddress, function seesaw.FunctionAddress, buf []byte, delay time.Duration) error

	// Write writes an entire array into a given module and function
	Write(module seesaw.ModuleBaseAddress, function seesaw.FunctionAddress, buf []byte) error
}

type Edge uint8

const (
	EdgeHigh Edge = iota
	EdgeLow
	EdgeFalling
	EdgeRising
)

// KeyEvent represents a pressed or released key
type KeyEvent uint8

func (k KeyEvent) Edge() Edge {
	return Edge(k & 0b11)
}

func (k KeyEvent) Key() uint8 {
	return uint8(k >> 2)
}

type SeesawKeypad struct {
	seesaw Seesaw
}

func New(dev Seesaw) *SeesawKeypad {
	return &SeesawKeypad{seesaw: dev}
}

// KeyEventCount returns the number of pending KeyEvent s in the FIFO queue
func (s *SeesawKeypad) KeyEventCount() (uint8, error) {
	//https://github.com/adafruit/Adafruit_Seesaw/blob/master/Adafruit_seesaw.cpp#L721
	buf := make([]byte, 1)
	err := s.seesaw.Read(seesaw.ModuleKeypadBase, seesaw.FunctionKeypadCount, buf, 500*time.Microsecond)
	return buf[0], err
}

// SetKeypadInterrupt enables or disables interrupts for key events
func (s *SeesawKeypad) SetKeypadInterrupt(enable bool) error {
	if enable {
		return s.seesaw.Write(seesaw.ModuleKeypadBase, seesaw.FunctionKeypadIntenset, []byte{0x1})
	}
	return s.seesaw.Write(seesaw.ModuleKeypadBase, seesaw.FunctionKeypadIntenclr, []byte{0x1})
}

// Read reads pending KeyEvent s from the FIFO
func (s *SeesawKeypad) Read(buf []KeyEvent) error {
	//https://github.com/adafruit/Adafruit_Seesaw/blob/master/Adafruit_seesaw.cpp#LL732C21-L732C21

	// use some unsafe magic to avoid copy-ing the entire buffer
	bytesBuf := *(*[]byte)(unsafe.Pointer(&buf))

	return s.seesaw.Read(seesaw.ModuleKeypadBase, seesaw.FunctionKeypadFifo, bytesBuf, 2*time.Millisecond)
}

// ConfigureKeypad enables or disables a key and edge on the keypad module
func (s *SeesawKeypad) ConfigureKeypad(key uint8, edge Edge, enable bool) error {

	//set STATE
	state := byte(0)
	if enable {
		state |= 0x01
	}

	//set ACTIVE
	state |= (1 << edge) << 1
	return s.seesaw.Write(seesaw.ModuleKeypadBase, seesaw.FunctionKeypadEvent, []byte{key, state})
}
