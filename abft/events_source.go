package abft

import (
	"github.com/metatechchain/meta-base/hash"
	"github.com/metatechchain/meta-base/inter/dag"
)

// EventSource is a callback for getting events from an external storage.
type EventSource interface {
	HasEvent(hash.Event) bool
	GetEvent(hash.Event) dag.Event
}
