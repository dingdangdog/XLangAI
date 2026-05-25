package llmchat

// Message 单轮对话消息（与 handler 层结构对应）。
type Message struct {
	Role    string
	Content string
}
