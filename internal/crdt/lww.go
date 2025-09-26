// internal/crdt/lww.go
package crdt

import "time"

// LWWRegister keeps the latest value by timestamp and tiebreaker (nodeID).
type LWWRegister struct {
	Value     any    `json:"value"`
	Timestamp int64  `json:"ts"`
	NodeID    string `json:"nodeId"`
	// optional wall-clock for debugging
	WallClock time.Time `json:"-"`
}

type LWWWrite struct {
	Value     any    `json:"value"`
	Timestamp int64  `json:"ts"`     // logical/lamport or HLC
	NodeID    string `json:"nodeId"` // tie-breaker
}

func (r *LWWRegister) Apply(op LWWWrite) {
	if r.Timestamp < op.Timestamp || (r.Timestamp == op.Timestamp && r.NodeID < op.NodeID) {
		r.Value = op.Value
		r.Timestamp = op.Timestamp
		r.NodeID = op.NodeID
	}
}

func (r *LWWRegister) Merge(other LWWRegister) LWWRegister {
	out := *r
	out.Apply(LWWWrite{
		Value:     other.Value,
		Timestamp: other.Timestamp,
		NodeID:    other.NodeID,
	})
	return out
}
