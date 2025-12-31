package main

import (
	"cmp"
	"machine"
	"math"
)

// empirical values
const (
	adcMin       = 256
	adcMax       = 65280
	adcThreshold = 64
)

type ADC struct {
	machine.ADC
	currentSetpoint int // midpoint of the sector in adjusted raw ADC value
	currentSector   int
	sectorCount     int
}

func newADC(sectorCount int) ADC {
	machine.InitADC()
	machineAdc := machine.ADC{
		Pin: machine.ADC0,
	}
	machineAdc.Configure(machine.ADCConfig{})
	return ADC{ADC: machineAdc, sectorCount: sectorCount, currentSetpoint: math.MaxInt /* trigger change on first fetch*/}
}

func (a *ADC) GetSector() (int, bool) {
	v := a.read()

	// hysteresis
	width := adcMax / a.sectorCount
	if abs(v-a.currentSetpoint) < adcThreshold+width/2 {
		return a.currentSector, false
	}

	// pick sector
	f := float32(v) / float32(adcMax)
	newSector := clamp(0, a.sectorCount-1, int(f*float32(a.sectorCount)))

	if a.currentSector == newSector {
		return a.currentSector, false
	}

	// updated sector
	a.currentSector = newSector
	a.currentSetpoint = newSector*width + width/2

	return a.currentSector, true
}

// read ADC and remove offset
func (a *ADC) read() int {
	v := int(a.Get())
	v -= adcMin
	return clamp(0, adcMax, v)
}

func clamp[T cmp.Ordered](a, b, v T) T {
	return max(min(v, b), a)
}

func abs[T ~int](a T) T {
	var zero T
	if a < zero {
		return -a
	}
	return a
}
