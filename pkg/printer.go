package pkg

import (
	"fmt"
	"math"
)

const (
	B float64 = 1 << (10 * iota)
	KiB
	MiB
	GiB
	TiB
	PiB
	EiB
	ZiB
	YiB
)

type unitWithText struct {
	unit float64
	text string
}

var units = []unitWithText{
	{
		B,
		"B",
	}, {
		KiB,
		"KiB",
	}, {
		MiB,
		"MiB",
	}, {
		GiB,
		"GiB",
	}, {
		TiB,
		"TiB",
	}, {
		PiB,
		"PiB",
	}, {
		EiB,
		"EiB",
	}, {
		ZiB,
		"ZiB",
	}, {
		YiB,
		"YiB",
	},
}

func PrettyPrintSize(size uint64) string {
	f := func(size float64, unit string) string { return fmt.Sprintf("%.2f %3s", size, unit) }
	if size == 0 {
		return f(0, "B")
	}
	canUse := int(math.Log(float64(size)) / math.Log(1024))
	if canUse >= len(units) {
		canUse = len(units) - 1
	}
	unit := units[canUse]
	return f(float64(size)/unit.unit, unit.text)
}
