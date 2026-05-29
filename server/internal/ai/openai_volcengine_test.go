package ai

import "testing"

func TestOpenAIChatCompletionsURL_VolcengineArk(t *testing.T) {
	base := "https://ark.cn-beijing.volces.com/api/v3"
	got := OpenAIChatCompletionsURL(base)
	want := "https://ark.cn-beijing.volces.com/api/v3/chat/completions"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestOpenAIChatCompletionsURL_OpenAI(t *testing.T) {
	got := OpenAIChatCompletionsURL("https://api.openai.com")
	want := "https://api.openai.com/v1/chat/completions"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestBuildChatCompletionTokenParams_ArkUsesMaxCompletionTokens(t *testing.T) {
	params := buildChatCompletionTokenParams("https://ark.cn-beijing.volces.com/api/v3", &ChatOptions{MaxTokens: 1024})
	if params["max_completion_tokens"] != 1024 {
		t.Fatalf("expected max_completion_tokens=1024 got %v", params)
	}
	if _, ok := params["max_tokens"]; ok {
		t.Fatal("ark should not use max_tokens")
	}
}
