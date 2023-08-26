package hi

const c1 = -8.784_694_755_56
const c2 = 1.611_394_11
const c3 = 2.338_548_838_89
const c4 = -0.146_116_05
const c5 = -0.012_308_094
const c6 = -0.016_424_827_7778
const c7 = 2.211_732e-3
const c8 = 7.2546e-4
const c9 = -3.582e-6

type HeatIndexEffect int

const (
	EffectUnknown HeatIndexEffect = iota
	EffectNone
	EffectCaution
	EffectExtremeCaution
	EffectDanger
	EffectExtremeDanger
)

func HeatIndexToEffect(index int32) HeatIndexEffect {
	if index > 54 {
		return EffectExtremeDanger
	}

	if index > 41 {
		return EffectDanger
	}

	if index > 32 {
		return EffectExtremeCaution
	}

	if index > 27 {
		return EffectCaution
	}

	return EffectNone
}

// Calculate calculates the heat index given temperature and relative humidity.
// 27–32 °C 	Caution: fatigue is possible with prolonged exposure and activity. Continuing activity could result in heat cramps.
// 32–41 °C 	Extreme caution: heat cramps and heat exhaustion are possible. Continuing activity could result in heat stroke.
// 41–54 °C 	Danger: heat cramps and heat exhaustion are likely; heat stroke is probable with continued activity.
// over 54 °C 	Extreme danger: heat stroke is imminent.
func Calculate(milliTemp, milliRH int32) int32 {
	if milliTemp < 0 {
		milliTemp = 0
	} else if milliTemp > 60*1000 {
		milliTemp = 60 * 1000
	}

	if milliRH < 0 {
		milliRH = 0
	} else if milliRH > 1000*100 {
		milliRH = 1000 * 100
	}

	T := float32(milliTemp) / 1000
	R := float32(milliRH) / 1000

	TxT := T * T
	RxR := R * R

	//https://en.wikipedia.org/wiki/Heat_index#Formula
	v := c1 + c2*T + c3*R + c4*T*R + c5*TxT + c6*RxR + c7*TxT*R + c8*T*RxR + c9*TxT*RxR

	return int32(v + 0.5)
}
