package stats

// DivideFloat64 returns the value of value divided by base.
// divide by zero case handled before calling.
func DivideFloat64(value, base float64) (float64, bool) {
	if base == 0 {
		return 0, false
	}
	return (value / base), true
}
