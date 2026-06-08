package protocol

import "testing"

func TestEncodeDecodeMessage(t *testing.T) {
	original := Message{
		Type:      TypeRequestStart,
		RequestID: NewRequestID(),
		Method:    "GET",
		Path:      "/users",
		Query:     "id=1",
		Target:    "http://127.0.0.1:3000",
		Headers:   map[string][]string{"X-Test": {"ok"}},
	}
	data, err := Encode(original)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != original.Type || got.RequestID != original.RequestID || got.Target != original.Target {
		t.Fatalf("decoded message mismatch: %#v", got)
	}
}

func TestPendingComplete(t *testing.T) {
	pending := NewPending[string]()
	ch := pending.Add("id")
	if !pending.Complete("id", "done") {
		t.Fatal("expected completion")
	}
	if got := <-ch; got != "done" {
		t.Fatalf("got %q", got)
	}
}
