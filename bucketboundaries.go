package histogram

func Range(start int64, stop int64, step int64) []int64 {
	// Step size  cannot be 0
	// If Step > 0, then start <= stop
	// If step < 0, then start >= stop
	if step == 0 {
		return nil
	} else if step > 0 && start > stop {
		return nil

	} else if step < 0 && start < stop {
		return nil
	}
	length := int((stop-start)/step + 1)
	values := make([]int64, length)
	value := start
	for i := 0; i < length; i++ {
		values[i] = value
		value += step
	}
	return values
}
