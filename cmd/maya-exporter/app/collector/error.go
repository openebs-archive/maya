package collector

type colErr struct {
	err error
}

func (e *colErr) Error() string {
	return e.err.Error()
}
