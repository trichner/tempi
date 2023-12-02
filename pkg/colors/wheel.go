package colors

import "image/color"

// ColorAt returns the color within an RGB wheel at the given position.
// This wraps around seamlessly, ideal for rainbows.
func ColorAt(p uint8, brightness uint8) color.RGBA {
	p = 255 - p
	if p < 85 {
		return color.RGBA{255 - p*3, 0, p * 3, brightness}
	}
	if p < 170 {
		p -= 85
		return color.RGBA{0, p * 3, 255 - p*3, brightness}
	}
	p -= 170
	return color.RGBA{p * 3, 255 - p*3, 0, brightness}
}

type Wheel struct {
	Brightness uint8
	pos        uint8
}

// Next increments the internal state of the color and returns the new RGBA
func (w *Wheel) Next() (c color.RGBA) {
	next := ColorAt(w.pos, w.Brightness)
	w.pos++
	return next
}
