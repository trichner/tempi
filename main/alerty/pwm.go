package main

import "machine"

type PWM interface {
	Set(channel uint8, value uint32)
	SetPeriod(period uint64) error
	Enable(bool)
	Top() uint32
	Configure(config machine.PWMConfig) error
	Channel(machine.Pin) (uint8, error)
}

func getPWM(pin machine.Pin) (PWM, uint8, error) {
	var pwms = [...]PWM{machine.PWM0, machine.PWM1, machine.PWM2, machine.PWM3, machine.PWM4, machine.PWM5, machine.PWM6, machine.PWM7}
	slice, err := machine.PWMPeripheral(pin)
	if err != nil {
		return nil, 0, err
	}
	pwm := pwms[slice]
	err = pwm.Configure(machine.PWMConfig{Period: 1e9 / 100}) // 100Hz for starters.
	if err != nil {
		return nil, 0, err
	}
	channel, err := pwm.Channel(pin)
	return pwm, channel, err
}
