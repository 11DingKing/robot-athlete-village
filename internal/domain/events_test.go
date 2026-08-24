package domain

import (
	"testing"
	"time"
)

func TestEventRoundTrip(t *testing.T) {
	e := NewEvent("e1", "stay.admitted", "stay", "7", time.Now(), map[string]string{"room": "A-101"})
	data, err := e.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	got, err := UnmarshalEvent(data)
	if err != nil || got.ID != e.ID || got.Payload["room"] != "A-101" {
		t.Fatalf("%+v %v", got, err)
	}
}
func TestEventValidation(t *testing.T) {
	base := Event{ID: "id", Kind: "kind", EntityType: "stay", EntityID: "1", OccurredAt: time.Now()}
	for _, e := range []Event{{}, {ID: base.ID, Kind: base.Kind, EntityType: base.EntityType, EntityID: base.EntityID}, {ID: "", Kind: base.Kind, EntityType: base.EntityType, EntityID: base.EntityID, OccurredAt: base.OccurredAt}} {
		if e.Validate() == nil {
			t.Fatalf("accepted %+v", e)
		}
	}
	if base.Validate() != nil {
		t.Fatal("base")
	}
}
func TestEventPayloadIsolation(t *testing.T) {
	e := NewEvent("e", "k", "t", "1", time.Now(), map[string]string{"a": "1"})
	copy := e.WithPayload("b", "2")
	copy.Payload["a"] = "changed"
	if e.Payload["a"] != "1" || copy.Payload["b"] != "2" {
		t.Fatal("shared payload")
	}
}
