// internal/crdt/types.go
package crdt

// CRDTType enumerates supported CRDTs.
type CRDTType string

const (
	TypeLWW   CRDTType = "lww"
	TypeORSet CRDTType = "orset"
)

// Version is a monotonically increasing integer per document.
type Version int64
