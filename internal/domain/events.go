package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

type Event struct {
	ID         string            `json:"id"`
	Kind       string            `json:"kind"`
	EntityType string            `json:"entity_type"`
	EntityID   string            `json:"entity_id"`
	OccurredAt time.Time         `json:"occurred_at"`
	Payload    map[string]string `json:"payload"`
}

func NewEvent(id, kind, entityType, entityID string, at time.Time, payload map[string]string) Event {
	return Event{ID: id, Kind: kind, EntityType: entityType, EntityID: entityID, OccurredAt: at.UTC(), Payload: clonePayload(payload)}
}
func (e Event) Validate() error {
	if e.ID == "" || e.Kind == "" || e.EntityType == "" || e.EntityID == "" {
		return fmt.Errorf("event identity required")
	}
	if e.OccurredAt.IsZero() {
		return fmt.Errorf("event time required")
	}
	return nil
}
func (e Event) Marshal() ([]byte, error) {
	if err := e.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(e)
}
func UnmarshalEvent(data []byte) (Event, error) {
	var e Event
	if err := json.Unmarshal(data, &e); err != nil {
		return e, err
	}
	if err := e.Validate(); err != nil {
		return e, err
	}
	e.Payload = clonePayload(e.Payload)
	return e, nil
}
func clonePayload(in map[string]string) map[string]string {
	if in == nil {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
func (e Event) WithPayload(key, value string) Event {
	copy := e
	copy.Payload = clonePayload(e.Payload)
	copy.Payload[key] = value
	return copy
}
