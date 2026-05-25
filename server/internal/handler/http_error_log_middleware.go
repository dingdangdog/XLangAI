package handler

import (
	"bytes"
	"encoding/json"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

// bodyLogWriter 捕获响应体，供错误日志提取 error/reason/code。
type bodyLogWriter struct {
	gin.ResponseWriter
	body bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	if len(b) > 0 {
		_, _ = w.body.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

func truncateLog(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

func extractAPIErrorSummary(body string) (errMsg, code, reason, detail string) {
	body = strings.TrimSpace(body)
	if body == "" {
		return "", "", "", ""
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		return truncateLog(body, 300), "", "", ""
	}
	if v, ok := m["error"].(string); ok {
		errMsg = v
	}
	if v, ok := m["code"].(string); ok {
		code = v
	}
	if v, ok := m["reason"].(string); ok {
		reason = v
	}
	if v, ok := m["detail"].(string); ok {
		detail = v
	}
	return errMsg, code, reason, detail
}

// HTTPErrorLogMiddleware 对 status>=400 写应用日志（与 Gin 访问日志互补）。
// 生产环境看容器 stdout 搜 [api] 即可，无需每个 handler 单独打日志。
func HTTPErrorLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		blw := &bodyLogWriter{ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		status := c.Writer.Status()
		if status < 400 {
			return
		}

		errMsg, code, reason, detail := extractAPIErrorSummary(blw.body.String())
		uid := strings.TrimSpace(c.GetString("user_id"))
		if uid == "" {
			uid = "-"
		}

		// 5xx 与配置类 4xx 同样记录；便于对照 [GIN] 行排查。
		log.Printf(
			"[api] %d %s %s user_id=%s client=%s query=%s code=%s reason=%s error=%s detail=%s",
			status,
			c.Request.Method,
			c.FullPath(),
			uid,
			c.ClientIP(),
			truncateLog(c.Request.URL.RawQuery, 120),
			truncateLog(code, 64),
			truncateLog(reason, 240),
			truncateLog(errMsg, 240),
			truncateLog(detail, 240),
		)
	}
}
