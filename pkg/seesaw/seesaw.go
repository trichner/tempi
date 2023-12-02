// Package seesaw provides a driver implementation to communicate with Adafruit's seesaw chip.
package seesaw

import (
	"errors"
	"strconv"
	"time"
)

const DefaultSeesawAddress = 0x49

// empirically determined standardDelay, the one from the official library seems to be too short (250us)
const defaultDelay = 100 * time.Millisecond

const (
	seesawHwIdCodeSAMD09  = 0x55 // HW ID code for SAMD09
	seesawHwIdCodeTINY8x7 = 0x87 // HW ID code for ATtiny817
)

type Seesaw interface {
	// Read reads a number of bytes from the device after sending the read command and waiting 'delay'. The delays depend
	// on the module and function and are documented in the seesaw datasheet
	Read(module ModuleBaseAddress, function FunctionAddress, buf []byte, delay time.Duration) error

	// Write writes an entire array into a given module and function
	Write(module ModuleBaseAddress, function FunctionAddress, buf []byte) error
}

type Device struct {
	bus           I2C
	Address       uint16
	hwid          byte
	standardDelay time.Duration
}

func New(bus I2C) *Device {
	return &Device{
		bus:           bus,
		Address:       DefaultSeesawAddress,
		standardDelay: defaultDelay,
	}
}

// SoftReset triggers a soft-reset of seesaw and waits for it to be ready
func (d *Device) SoftReset() error {
	err := d.WriteRegister(ModuleStatusBase, FunctionStatusSwrst, 0xFF)
	if err != nil {
		return errors.New("failed sending soft-reset command: " + err.Error())
	}

	return d.waitForReset()
}

func (d *Device) waitForReset() error {
	// give the device a little bit of time to reset
	time.Sleep(time.Second)

	var lastErr error
	tries := 0
	for ; tries < 20; tries++ {
		hwid, err := d.readHardwareID()
		if err == nil {
			d.hwid = hwid
			return nil
		}
		lastErr = err
		time.Sleep(20 * time.Millisecond)
	}
	return errors.New("failed to wait for device to start: " + lastErr.Error())
}

func (d *Device) readHardwareID() (byte, error) {
	hwid, err := d.ReadRegister(ModuleStatusBase, FunctionStatusHwId)
	if err != nil {
		return 0, err
	}

	if hwid == seesawHwIdCodeSAMD09 || hwid == seesawHwIdCodeTINY8x7 {
		return hwid, nil
	}

	return 0, errors.New("unknown hardware ID: " + strconv.FormatUint(uint64(hwid), 16))
}

// WriteRegister writes a single seesaw register
func (d *Device) WriteRegister(module ModuleBaseAddress, function FunctionAddress, value byte) error {
	buf := []byte{byte(module), byte(function), value}
	return d.bus.Tx(d.Address, buf, nil)
}

// ReadRegister reads a single register from seesaw
func (d *Device) ReadRegister(module ModuleBaseAddress, function FunctionAddress) (byte, error) {
	buf := make([]byte, 1)
	err := d.Read(module, function, buf, d.standardDelay)
	if err != nil {
		return 0, err
	}
	return buf[0], nil
}

// Read reads a number of bytes from the device after sending the read command and waiting 'standardDelay'. The delays depend
// on the module and function and are documented in the seesaw datasheet
func (d *Device) Read(module ModuleBaseAddress, function FunctionAddress, buf []byte, delay time.Duration) error {
	prefix := []byte{byte(module), byte(function)}
	err := d.bus.Tx(d.Address, prefix, nil)
	if err != nil {
		return err
	}

	// This is needed for the client seesaw device to flush its RX buffer and process the command.
	// See seesaw datasheet for timings for specific modules.
	time.Sleep(delay)

	return d.bus.Tx(d.Address, nil, buf)
}

// Write writes an entire array into a given module and function
func (d *Device) Write(module ModuleBaseAddress, function FunctionAddress, buf []byte) error {
	prefix := []byte{byte(module), byte(function)}
	data := append(prefix, buf...)
	return d.bus.Tx(d.Address, data, nil)
}

var hexLookupTable = [16]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'}

func byteToHexString(s byte) string {
	formatted := make([]byte, 2)
	formatted[0] = hexLookupTable[s>>4]
	formatted[1] = hexLookupTable[s&0xf]
	return string(formatted)
}
