package proxy

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/dev2k6/command-code-proxy-server/internal/api"
)

// Upstream rejects messages whose content serializes as null.
// Empty-string / nil content must become a valid text part instead.
func TestConvertMessagesNeverEmitsNullContent(t *testing.T) {
	msgs := []api.OpenAIMessage{
		{Role: "assistant", Content: ""},                            // empty string
		{Role: "assistant", Content: nil},                           // explicit null
		{Role: "user", Content: ""},                                 // user empty
		{Role: "assistant", ToolCalls: []api.ToolCall{{              // tool call w/ empty content
			ID:       "call_1",
			Type:     "function",
			Function: api.FunctionCall{Name: "f", Arguments: "{}"},
		}}},
	}

	cc := ConvertMessages(msgs)
	body := api.CCRequestBody{
		Params: api.CCChatParams{Model: "m", Messages: cc},
	}
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	s := string(data)
	if strings.Contains(s, "\"content\":null") {
		t.Fatalf("serialized body contains \"content\":null:\n%s", s)
	}
	for i, m := range cc {
		if m.Content == nil || len(m.Content) == 0 {
			t.Fatalf("message %d (%s) has empty content parts", i, m.Role)
		}
	}
}
