package toggler

import "machine"

type Toggler struct {
	pin   machine.Pin
	count uint8
}

func SetupToggler(pin machine.Pin) *Toggler {
	pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	return &Toggler{pin: pin}
}

func (t *Toggler) Toggle() {
	if t.count%2 == 0 {
		t.pin.High()
	} else {
		t.pin.Low()
	}
	t.count++
}
