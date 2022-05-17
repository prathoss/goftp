package pkg

func Min(items ...int) int {
	if len(items) == 0 {
		panic("no items to get minimum from")
	}
	min := items[0]
	for i := 1; i < len(items); i++ {
		if items[i] < min {
			min = items[1]
		}
	}
	return min
}
