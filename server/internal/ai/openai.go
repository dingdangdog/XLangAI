package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

type Client struct {
	apiKey  string
	model   string
	baseURL string
}

// NormalizeOpenAIBaseURL 去掉末尾 / 以及重复的「/v1」后缀，再与 /v1/chat/...、/v1/audio/... 拼接。
// 若 Base URL 配成 https://api.xxx.com/v1，不处理会得到 /v1/v1/... 从而 404。
func NormalizeOpenAIBaseURL(baseURL string) string {
	s := strings.TrimSpace(baseURL)
	if s == "" {
		return "https://api.openai.com"
	}
	s = strings.TrimRight(s, "/")
	const suf = "/v1"
	for strings.HasSuffix(s, suf) {
		s = strings.TrimSuffix(s, suf)
		s = strings.TrimRight(s, "/")
	}
	return s
}

func NewOpenAIClient(apiKey, model, baseURL string) *Client {
	if model == "" {
		model = "gpt-4o-mini"
	}
	base := NormalizeOpenAIBaseURL(baseURL)
	return &Client{apiKey: apiKey, model: model, baseURL: base}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

// ChatUsage OpenAI chat.completions 返回的 usage 字段。
type ChatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Usage *ChatUsage `json:"usage"`
}

func (c *Client) Chat(ctx context.Context, systemPrompt string, messages []struct{ Role, Content string }) (string, *ChatUsage, error) {
	msgs := make([]chatMessage, 0, len(messages)+1)
	msgs = append(msgs, chatMessage{Role: "system", Content: systemPrompt})
	for _, m := range messages {
		msgs = append(msgs, chatMessage{Role: m.Role, Content: m.Content})
	}

	body, _ := json.Marshal(chatRequest{Model: c.model, Messages: msgs})
	url := strings.TrimRight(c.baseURL, "/") + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", nil, fmt.Errorf("openai api error %d: %s (POST %s)", resp.StatusCode, string(b), url)
	}

	var r chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", nil, err
	}
	if len(r.Choices) == 0 {
		return "", nil, fmt.Errorf("no response from openai")
	}
	return r.Choices[0].Message.Content, r.Usage, nil
}

type transcriptionResponse struct {
	Text string `json:"text"`
}

func (c *Client) Transcribe(ctx context.Context, filename string, data []byte, model string, language *string) (string, error) {
	if model == "" {
		model = "whisper-1"
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("model", model); err != nil {
		return "", err
	}
	if language != nil {
		lang := strings.TrimSpace(*language)
		if lang != "" {
			if err := writer.WriteField("language", lang); err != nil {
				return "", err
			}
		}
	}

	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return "", err
	}
	if _, err := part.Write(data); err != nil {
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", err
	}

	url := strings.TrimRight(c.baseURL, "/") + "/v1/audio/transcriptions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openai transcription error %d: %s (POST %s)", resp.StatusCode, string(b), url)
	}

	var r transcriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", err
	}
	if r.Text == "" {
		return "", fmt.Errorf("empty transcription")
	}
	return r.Text, nil
}
