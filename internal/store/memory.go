// internal/store/memory.go
package store

import (
	"errors"
	"sync"

	"github.com/oxtx/go-crdt-sync/internal/crdt"
)

type Doc struct {
	Type    crdt.CRDTType `json:"type"`
	Version crdt.Version  `json:"version"`

	LWW   crdt.LWWRegister `json:"lww,omitempty"`
	ORSet crdt.ORSet       `json:"orset,omitempty"`
}

type Operation struct {
	Seq crdt.Version   `json:"seq"`
	Op  map[string]any `json:"op"`
}

type MemoryStore struct {
	mu   sync.RWMutex
	docs map[string]*Doc
	ops  map[string][]Operation
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		docs: make(map[string]*Doc),
		ops:  make(map[string][]Operation),
	}
}

func (m *MemoryStore) GetDoc(id string) (*Doc, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	d, ok := m.docs[id]
	if !ok {
		return nil, errors.New("not found")
	}
	// shallow copy to avoid external mutation
	cp := *d
	return &cp, nil
}

func (m *MemoryStore) PutDoc(id string, d Doc) (Doc, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	d.Version = 0
	switch d.Type {
	case crdt.TypeLWW:
		// keep provided LWW snapshot as is
	case crdt.TypeORSet:
		if d.ORSet.Elements == nil {
			d.ORSet = crdt.NewORSet()
		}
	default:
		return Doc{}, errors.New("unsupported type")
	}
	m.docs[id] = &d
	m.ops[id] = nil
	return d, nil
}

func (m *MemoryStore) AppendOps(id string, incoming []Operation) (Doc, []Operation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	doc, ok := m.docs[id]
	if !ok {
		return Doc{}, nil, errors.New("not found")
	}

	for _, op := range incoming {
		doc.Version++
		op.Seq = doc.Version

		switch doc.Type {
		case crdt.TypeLWW:
			// op.Op = { "lww_write": { value, ts, nodeId } }
			if raw, ok := op.Op["lww_write"].(map[string]any); ok {
				write := crdt.LWWWrite{
					Value:     raw["value"],
					Timestamp: int64(coerceFloat(raw["ts"])),
					NodeID:    coerceString(raw["nodeId"]),
				}
				doc.LWW.Apply(write)
			}
		case crdt.TypeORSet:
			// "orset_add": {item, tag} OR "orset_remove": {item, seen:[]}
			if raw, ok := op.Op["orset_add"].(map[string]any); ok {
				doc.ORSet.ApplyAdd(crdt.ORSetAdd{
					Item: coerceString(raw["item"]),
					Tag:  coerceString(raw["tag"]),
				})
			}
			if raw, ok := op.Op["orset_remove"].(map[string]any); ok {
				seen := coerceStringSlice(raw["seen"])
				doc.ORSet.ApplyRemove(crdt.ORSetRemove{
					Item: coerceString(raw["item"]),
					Seen: seen,
				})
			}
		}
		m.ops[id] = append(m.ops[id], op)
	}
	// return current doc and all ops (caller may filter by since)
	cp := *doc
	return cp, append([]Operation(nil), m.ops[id]...), nil
}

func (m *MemoryStore) GetOpsSince(id string, since crdt.Version) ([]Operation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	all := m.ops[id]
	if all == nil {
		return nil, errors.New("not found")
	}
	var out []Operation
	for _, op := range all {
		if op.Seq > since {
			out = append(out, op)
		}
	}
	return out, nil
}

func coerceFloat(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case jsonNumber:
		f, _ := t.Float64()
		return f
	default:
		return 0
	}
}

func coerceString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func coerceStringSlice(v any) []string {
	out := []string{}
	if v == nil {
		return out
	}
	if arr, ok := v.([]any); ok {
		for _, it := range arr {
			if s, ok := it.(string); ok {
				out = append(out, s)
			}
		}
	}
	return out
}

// json.Number without importing encoding/json here.
type jsonNumber interface {
	Float64() (float64, error)
}
