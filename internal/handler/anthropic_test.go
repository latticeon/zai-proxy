package handler

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"zai-proxy/internal/model"
)

func TestAnthropicNonStreamResponse_StripsSeparatorRuleOutput(t *testing.T) {
	body := newFakeBody(
		sseEvent("answer", "你¿好", ""),
		sseEventDone(),
	)

	w := httptest.NewRecorder()
	handleAnthropicNonStreamWithSeparatorRule(w, body, "msg_test", "glm-4.7", "claude-sonnet-4-6", nil, true)

	var resp model.AnthropicResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(resp.Content) != 1 {
		t.Fatalf("len(Content) = %d, want 1", len(resp.Content))
	}
	if resp.Content[0].Text != "你好" {
		t.Fatalf("Text = %q, want %q", resp.Content[0].Text, "你好")
	}
}
