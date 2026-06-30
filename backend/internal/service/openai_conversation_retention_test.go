package service

import "testing"

func TestBuildOpenAIConversationSessionKeyUsesStableClientSession(t *testing.T) {
	key1 := BuildOpenAIConversationSessionKey(7, "session-abc", "req-1")
	key2 := BuildOpenAIConversationSessionKey(7, "session-abc", "req-2")
	if key1 != key2 {
		t.Fatalf("expected same client session to map to same key, got %q != %q", key1, key2)
	}
}

func TestBuildOpenAIConversationSessionKeyFallsBackToRequestID(t *testing.T) {
	key1 := BuildOpenAIConversationSessionKey(7, "", "req-1")
	key2 := BuildOpenAIConversationSessionKey(7, "", "req-2")
	if key1 == key2 {
		t.Fatalf("expected fallback request ids to differ when client session is absent")
	}
}

func TestShouldRecordOpenAIConversationRetention(t *testing.T) {
	cases := map[string]bool{
		"/v1/responses":                  true,
		"/responses":                     true,
		"/backend-api/codex/responses":   true,
		"/v1/chat/completions":           true,
		"/chat/completions":              true,
		"/openai/v1/chat/completions":    true,
		"/v1/embeddings":                 false,
		"/v1/messages":                   false,
		"/responses-fake":                false,
		"/backend-api/codex/other":       false,
	}
	for input, want := range cases {
		if got := shouldRecordOpenAIConversationRetention(input); got != want {
			t.Fatalf("shouldRecordOpenAIConversationRetention(%q) = %v, want %v", input, got, want)
		}
	}
}
