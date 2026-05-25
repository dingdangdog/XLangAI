package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// azureMaxSSMLBody 单次请求 SSML 长度保守上限（含标签）；超长则拆段后拼接 mp3。
const azureMaxSSMLBody = 1800

type azureExtraConfig struct {
	OutputFormat string `json:"output_format"`
	Region       string `json:"region"`
}

// AzureTTSSpeechREST 使用 Azure Cognitive Services REST（与 jsdemo/tts/azure.js 中 key+region+voice 一致，输出为 mp3）。
func AzureTTSSpeechREST(ctx context.Context, subscriptionKey, region, outputFormat, azureVoiceName, text string) ([]byte, error) {
	if subscriptionKey == "" || region == "" {
		return nil, fmt.Errorf("azure tts: missing subscription key or region")
	}
	if azureVoiceName == "" {
		return nil, fmt.Errorf("azure tts: empty voice name")
	}
	region = strings.TrimSpace(strings.ToLower(region))
	endpoint := fmt.Sprintf("https://%s.tts.speech.microsoft.com/cognitiveservices/v1", region)
	if outputFormat == "" {
		outputFormat = "audio-16khz-128kbitrate-mono-mp3"
	}

	chunks := splitTextForAzureTTS(text, azureMaxSSMLBody)
	var out [][]byte
	for _, chunk := range chunks {
		ssml := buildAzureSSML(azureVoiceName, chunk)
		b, err := azureSingleRequest(ctx, endpoint, subscriptionKey, outputFormat, ssml)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	if len(out) == 1 {
		return out[0], nil
	}
	var buf bytes.Buffer
	for _, b := range out {
		buf.Write(b)
	}
	return buf.Bytes(), nil
}

func azureSingleRequest(ctx context.Context, endpoint, key, outputFormat, ssml string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(ssml))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Ocp-Apim-Subscription-Key", key)
	req.Header.Set("Content-Type", "application/ssml+xml")
	req.Header.Set("X-Microsoft-OutputFormat", outputFormat)
	req.Header.Set("User-Agent", "wlltalk-server")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("azure tts error %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

func buildAzureSSML(voiceName, plainText string) string {
	lang := localeFromAzureVoice(voiceName)
	var esc strings.Builder
	_ = xml.EscapeText(&esc, []byte(plainText))
	return fmt.Sprintf(
		"<speak version='1.0' xml:lang='%s'><voice xml:lang='%s' name='%s'>%s</voice></speak>",
		lang, lang, xmlEscapeAttr(voiceName), esc.String(),
	)
}

func xmlEscapeAttr(s string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}

func localeFromAzureVoice(v string) string {
	parts := strings.Split(v, "-")
	if len(parts) >= 2 && len(parts[0]) >= 2 && len(parts[1]) >= 2 {
		return strings.ToLower(parts[0]) + "-" + strings.ToUpper(parts[1])
	}
	return "en-US"
}

func splitTextForAzureTTS(text string, maxRunes int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return []string{text}
	}
	var out []string
	for start := 0; start < len(runes); {
		end := start + maxRunes
		if end > len(runes) {
			end = len(runes)
		} else {
			// 尽量在标点处断开
			slice := runes[start:end]
			cut := -1
			for i := len(slice) - 1; i > len(slice)/2; i-- {
				switch slice[i] {
				case '。', '！', '？', '.', '!', '?', '；', ';', '，', ',', '\n':
					cut = i + 1
				}
				if cut > 0 {
					break
				}
			}
			if cut > 0 {
				end = start + cut
			}
		}
		out = append(out, string(runes[start:end]))
		start = end
	}
	return out
}

// ParseAzureOutputFormat 从 config JSON 读取 output_format；可额外覆盖 region（列优先）。
func ParseAzureOutputFormat(configJSON string) (outputFormat string, regionFromJSON string) {
	var ex azureExtraConfig
	if configJSON != "" {
		_ = json.Unmarshal([]byte(configJSON), &ex)
	}
	return strings.TrimSpace(ex.OutputFormat), strings.TrimSpace(ex.Region)
}
