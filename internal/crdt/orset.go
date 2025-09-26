// internal/crdt/orset.go
package crdt

// OR-Set with add-tags (unique IDs) per element.
type ORSet struct {
	// elements maps item -> set of tag IDs that are currently present.
	Elements map[string]map[string]struct{} `json:"elements"`
}

type ORSetAdd struct {
	Item string `json:"item"`
	Tag  string `json:"tag"` // unique per add
}

type ORSetRemove struct {
	Item string   `json:"item"`
	Seen []string `json:"seen"` // tags observed by remover
}

func NewORSet() ORSet {
	return ORSet{Elements: map[string]map[string]struct{}{}}
}

func (s *ORSet) ApplyAdd(op ORSetAdd) {
	if s.Elements[op.Item] == nil {
		s.Elements[op.Item] = map[string]struct{}{}
	}
	s.Elements[op.Item][op.Tag] = struct{}{}
}

func (s *ORSet) ApplyRemove(op ORSetRemove) {
	tags := s.Elements[op.Item]
	if tags == nil {
		return
	}
	for _, t := range op.Seen {
		delete(tags, t)
	}
	if len(tags) == 0 {
		delete(s.Elements, op.Item)
	}
}

func (s ORSet) Merge(other ORSet) ORSet {
	out := NewORSet()
	// union of tags for each item
	for item, tags := range s.Elements {
		if out.Elements[item] == nil {
			out.Elements[item] = map[string]struct{}{}
		}
		for t := range tags {
			out.Elements[item][t] = struct{}{}
		}
	}
	for item, tags := range other.Elements {
		if out.Elements[item] == nil {
			out.Elements[item] = map[string]struct{}{}
		}
		for t := range tags {
			out.Elements[item][t] = struct{}{}
		}
	}
	return out
}

// Items returns the set of visible items.
func (s ORSet) Items() []string {
	out := make([]string, 0, len(s.Elements))
	for item := range s.Elements {
		out = append(out, item)
	}
	return out
}
