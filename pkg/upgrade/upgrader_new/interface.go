package upgrader

// Upgrader abstracts the upgrade of a resource
type Upgrader interface {
	Upgrade() error
}
