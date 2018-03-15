package crdops

const (
	// SuccessSynced is used as part of the Event 'reason' when a spc is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a spc fails
	// to sync due to a spc of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a spc already existing
	MessageResourceExists = "Resource %q already exists and is not managed by spc"
	// MessageResourceSynced is the message used for an Event fired when a spc
	// is synced successfully
	MessageResourceSynced = "Resource synced successfully"
)

// QueueLoad is for storing the key and type of operation before entering workqueue
type QueueLoad struct {
	key       string
	operation string
}
