// internal/service/service.go
package service

import (
	"errors"
	"time"

	"github.com/oxtx/go-crdt-sync/internal/crdt"
	"github.com/oxtx/go-crdt-sync/internal/store"
)

type PutDocRequest struct {
	Type     crdt.CRDTType `json:"type"`
	Snapshot any           `json:"snapshot,omitempty"`
}

type PutDocResponse struct {
	Version  crdt.Version  `json:"version"`
	Type     crdt.CRDTType `json:"type"`
	Snapshot any           `json:"snapshot"`
}

type PostOpsRequest struct {
	Since crdt.Version     `json:"since"`
	Ops   []map[string]any `json:"ops"`
}

type PostOpsResponse struct {
	Version  crdt.Version `json:"version"`
	Snapshot any          `json:"snapshot"`
}

type GetDocResponse struct {
	Version  crdt.Version  `json:"version"`
	Type     crdt.CRDTType `json:"type"`
	Snapshot any           `json:"snapshot"`
}

type Store interface {
	GetDoc(id string) (*store.Doc, error)
	PutDoc(id string, d store.Doc) (store.Doc, error)
	AppendOps(id string, incoming []store.Operation) (store.Doc, []store.Operation, error)
	GetOpsSince(id string, since crdt.Version) ([]store.Operation, error)
}

type Service struct {
	st Store
}

func New(st Store) *Service {
	return &Service{st: st}
}

func (s *Service) PutDoc(docID string, req PutDocRequest) (PutDocResponse, error) {
	d := store.Doc{Type: req.Type}
	switch req.Type {
	case crdt.TypeLWW:
		// Snapshot can be {value, ts, nodeId} or just value (we’ll wrap it)
		if snap, ok := req.Snapshot.(map[string]any); ok {
			d.LWW.Value = snap["value"]
			d.LWW.NodeID = str(snap["nodeId"])
			d.LWW.Timestamp = i64(snap["ts"])
		} else {
			// accept raw value, assign ts=now, nodeId="server"
			d.LWW.Value = req.Snapshot
			d.LWW.NodeID = "server"
			d.LWW.Timestamp = time.Now().UnixNano()
		}
	case crdt.TypeORSet:
		d.ORSet = crdt.NewORSet()
		// Optionally allow initial items: snapshot = []string
		if arr, ok := req.Snapshot.([]any); ok {
			for i, _ := range arr {
				if s, ok := arr[i].(string); ok {
					d.ORSet.ApplyAdd(crdt.ORSetAdd{Item: s, Tag: "init:" + s})
				}
			}
		}
	default:
		return PutDocResponse{}, errors.New("unsupported type")
	}
	doc, err := s.st.PutDoc(docID, d)
	if err != nil {
		return PutDocResponse{}, err
	}
	return PutDocResponse{
		Version:  doc.Version,
		Type:     doc.Type,
		Snapshot: snapshotOf(doc),
	}, nil
}

func (s *Service) GetDoc(docID string) (GetDocResponse, error) {
	doc, err := s.st.GetDoc(docID)
	if err != nil {
		return GetDocResponse{}, err
	}
	return GetDocResponse{
		Version:  doc.Version,
		Type:     doc.Type,
		Snapshot: snapshotOf(*doc),
	}, nil
}

func (s *Service) PostOps(docID string, req PostOpsRequest) (PostOpsResponse, error) {
	// In a fuller implementation, we’d filter/validate ops and ensure causal order.
	incoming := make([]store.Operation, len(req.Ops))
	for i, op := range req.Ops {
		incoming[i] = store.Operation{Op: op}
	}
	doc, _, err := s.st.AppendOps(docID, incoming)
	if err != nil {
		return PostOpsResponse{}, err
	}
	return PostOpsResponse{
		Version:  doc.Version,
		Snapshot: snapshotOf(doc),
	}, nil
}

func snapshotOf(doc store.Doc) any {
	switch doc.Type {
	case crdt.TypeLWW:
		return map[string]any{
			"value":  doc.LWW.Value,
			"ts":     doc.LWW.Timestamp,
			"nodeId": doc.LWW.NodeID,
		}
	case crdt.TypeORSet:
		// return list of items
		return doc.ORSet.Items()
	default:
		return nil
	}
}

func str(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func i64(v any) int64 {
	switch t := v.(type) {
	case int64:
		return t
	case int:
		return int64(t)
	case float64:
		return int64(t)
	default:
		return 0
	}
}
