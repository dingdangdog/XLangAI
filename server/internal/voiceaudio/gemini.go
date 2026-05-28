package voiceaudio

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"xlangai/server/internal/ai"
	"xlangai/server/internal/llmchat"
)

func geminiGenerate(ctx context.Context, in Input, audioOut bool) (*Output, error) {
	key := strings.TrimSpace(in.APIKey)
	if key == "" {
		return nil, ErrProviderNotReady
	}
	base := strings.TrimSpace(in.BaseURL)
	if base == "" {
		base = "https://generativelanguage.googleapis.com"
	}
	base = strings.TrimRight(base, "/")
	model := strings.TrimSpace(in.ModelCode)
	if model == "" {
		if audioOut {
			model = "gemini-2.5-flash-preview-tts"
		} else {
			model = "gemini-2.0-flash"
		}
	}
	model = strings.TrimPrefix(model, "models/")

	extra := parseRoleConfig(in.RoleConfig)
	voice := geminiVoiceName(in.VoiceCode)

	type part struct {
		Text       string `json:"text,omitempty"`
		InlineData *struct {
			MimeType string `json:"mimeType"`
			Data     string `json:"data"`
		} `json:"inlineData,omitempty"`
	}
	type content struct {
		Role  string `json:"role"`
		Parts []part `json:"parts"`
	}

	contents := make([]content, 0, len(in.Messages)+1)
	for _, m := range in.Messages {
		role := strings.ToLower(strings.TrimSpace(m.Role))
		switch role {
		case "assistant":
			role = "model"
		case "user":
			role = "user"
		default:
			continue
		}
		if strings.TrimSpace(m.Content) == "" {
			continue
		}
		contents = append(contents, content{
			Role:  role,
			Parts: []part{{Text: m.Content}},
		})
	}

	userParts := make([]part, 0, 2)
	if len(in.UserAudio) > 0 {
		mime := strings.TrimSpace(in.UserAudioMime)
		if mime == "" {
			mime = "audio/webm"
		}
		userParts = append(userParts, part{
			InlineData: &struct {
				MimeType string `json:"mimeType"`
				Data     string `json:"data"`
			}{
				MimeType: mime,
				Data:     base64.StdEncoding.EncodeToString(in.UserAudio),
			},
		})
	}
	if t := strings.TrimSpace(in.UserText); t != "" {
		userParts = append(userParts, part{Text: t})
	}
	if len(userParts) == 0 {
		return nil, fmt.Errorf("%w: empty user input", ErrFailed)
	}
	contents = append(contents, content{Role: "user", Parts: userParts})

	genCfg := map[string]any{}
	if audioOut {
		genCfg["responseModalities"] = []string{"AUDIO", "TEXT"}
		if len(extra.ResponseModalities) > 0 {
			genCfg["responseModalities"] = extra.ResponseModalities
		}
		genCfg["speechConfig"] = map[string]any{
			"voiceConfig": map[string]any{
				"prebuiltVoiceConfig": map[string]string{
					"voiceName": voice,
				},
			},
		}
	} else {
		genCfg["responseModalities"] = []string{"TEXT"}
	}

	reqBody := map[string]any{
		"contents":         contents,
		"generationConfig": genCfg,
	}
	if sp := strings.TrimSpace(in.SystemPrompt); sp != "" {
		reqBody["systemInstruction"] = map[string]any{
			"parts": []part{{Text: sp}},
		}
	}

	body, _ := json.Marshal(reqBody)
	reqURL := fmt.Sprintf("%s/v1beta/models/%s:generateContent", base, model)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", key)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: gemini %d %s", ErrFailed, resp.StatusCode, truncate(b, 400))
	}

	var parsed struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text       string `json:"text"`
					InlineData *struct {
						MimeType string `json:"mimeType"`
						Data     string `json:"data"`
					} `json:"inlineData"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return nil, err
	}

	out := &Output{}
	if len(parsed.Candidates) > 0 {
		for _, p := range parsed.Candidates[0].Content.Parts {
			if strings.TrimSpace(p.Text) != "" {
				out.Text += p.Text
			}
			if p.InlineData != nil && p.InlineData.Data != "" {
				raw, err := base64.StdEncoding.DecodeString(p.InlineData.Data)
				if err != nil {
					return nil, err
				}
				out.Audio = raw
				out.AudioMime = strings.TrimSpace(p.InlineData.MimeType)
				if out.AudioMime == "" {
					out.AudioMime = "audio/mpeg"
				}
			}
		}
	}
	out.Text = strings.TrimSpace(out.Text)
	out.Usage = &ai.ChatUsage{
		PromptTokens:     parsed.UsageMetadata.PromptTokenCount,
		CompletionTokens: parsed.UsageMetadata.CandidatesTokenCount,
		TotalTokens:      parsed.UsageMetadata.TotalTokenCount,
	}
	if out.Usage.TotalTokens == 0 {
		out.Usage.TotalTokens = out.Usage.PromptTokens + out.Usage.CompletionTokens
	}
	return out, nil
}

func truncate(b []byte, max int) string {
	s := strings.TrimSpace(string(b))
	if len(s) > max {
		return s[:max] + "…"
	}
	return s
}

func geminiTextChat(ctx context.Context, in Input) (*Output, error) {
	key := strings.TrimSpace(in.APIKey)
	if key == "" {
		return nil, ErrProviderNotReady
	}
	llmIn := llmchat.ServiceInputFromRepo("gemini", in.BaseURL, key, in.ModelCode, in.Config)
	text, usage, err := llmchat.Chat(ctx, llmIn, in.SystemPrompt, in.Messages)
	if err != nil {
		return nil, err
	}
	return &Output{Text: text, Usage: usage}, nil
}
