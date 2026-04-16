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

func TestAnthropicNonStreamResponse_AppendsFallbackContentFromTerminalPacket(t *testing.T) {
	body := newFakeBody(
		sseEvent("answer", "第八天", ""),
		`data: {"type":"chat:completion","data":{"content":"非常抱歉，我目前无法提供你需要的具体信息。","done":true,"error":{"code":"SENSITIVE","detail":"非常抱歉，我目前无法提供你需要的具体信息。"}}}`,
		sseEventDone(),
	)

	w := httptest.NewRecorder()
	handleAnthropicNonStream(w, body, "msg_test", "glm-4.7", "claude-sonnet-4-6", nil)

	var resp model.AnthropicResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(resp.Content) != 1 {
		t.Fatalf("len(Content) = %d, want 1", len(resp.Content))
	}
	want := "第八天非常抱歉，我目前无法提供你需要的具体信息。"
	if resp.Content[0].Text != want {
		t.Fatalf("Text = %q, want %q", resp.Content[0].Text, want)
	}
}
