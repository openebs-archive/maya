package patch

// Patcher abstracts the patching of components
type Patcher interface {
	PreChecks(from, to string) error
	Patch(from, to string) error
	// TODO
	// Validate() error
}
