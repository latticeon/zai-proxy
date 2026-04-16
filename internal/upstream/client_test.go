package upstream

import "testing"

func TestInsertMessageBeforeLast_Empty(t *testing.T) {
	msg := map[string]interface{}{"role": "user", "content": "rule"}
	got := insertMessageBeforeLast(nil, msg)

	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1", len(got))
	}
	if got[0]["content"] != "rule" {
		t.Fatalf("got[0].content = %v, want rule", got[0]["content"])
	}
}

func TestInsertMessageBeforeLast_SingleMessage(t *testing.T) {
	rule := map[string]interface{}{"role": "user", "content": "rule"}
	last := map[string]interface{}{"role": "user", "content": "question"}

	got := insertMessageBeforeLast([]map[string]interface{}{last}, rule)

	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}
	if got[0]["content"] != "rule" {
		t.Fatalf("got[0].content = %v, want rule", got[0]["content"])
	}
	if got[1]["content"] != "question" {
		t.Fatalf("got[1].content = %v, want question", got[1]["content"])
	}
}

func TestInsertMessageBeforeLast_MultipleMessages(t *testing.T) {
	first := map[string]interface{}{"role": "user", "content": "history"}
	rule := map[string]interface{}{"role": "user", "content": "rule"}
	last := map[string]interface{}{"role": "user", "content": "question"}

	got := insertMessageBeforeLast([]map[string]interface{}{first, last}, rule)

	if len(got) != 3 {
		t.Fatalf("len(got) = %d, want 3", len(got))
	}
	if got[0]["content"] != "history" {
		t.Fatalf("got[0].content = %v, want history", got[0]["content"])
	}
	if got[1]["content"] != "rule" {
		t.Fatalf("got[1].content = %v, want rule", got[1]["content"])
	}
	if got[2]["content"] != "question" {
		t.Fatalf("got[2].content = %v, want question", got[2]["content"])
	}
}
