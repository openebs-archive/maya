package v1alpha1

type CollectorError struct {
	Err error
}

func (c *CollectorError) Error() string {
	return c.Err.Error()
}
