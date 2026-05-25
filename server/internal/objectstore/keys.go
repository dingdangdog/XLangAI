package objectstore

import (
	"strings"

	"github.com/google/uuid"
)

// Category 决定 R2/OSS 对象键前缀（与存储桶内目录一致）。
type Category string

const (
	CategoryAvatar    Category = "useravatar"
	CategoryUserAudio Category = "useraudio"
	CategoryAIAudio   Category = "aiaudio"
)

// BuildObjectKey 生成对象键，例如 useraudio/{uuid}.webm。
func BuildObjectKey(cat Category, ext string) string {
	return string(cat) + "/" + uuid.New().String() + normalizeExt(ext)
}

// KeyFromPublicURL 从公网 URL 解析对象键；不匹配时返回空字符串。
func KeyFromPublicURL(publicBaseURL, objectURL string) string {
	base := strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
	u := strings.TrimSpace(objectURL)
	if base == "" || u == "" {
		return ""
	}
	prefix := base + "/"
	if !strings.HasPrefix(u, prefix) {
		if strings.HasPrefix(base, "https://") || strings.HasPrefix(base, "http://") {
			return ""
		}
		prefix = "https://" + base + "/"
		if !strings.HasPrefix(u, prefix) {
			return ""
		}
	}
	return strings.TrimPrefix(u, prefix)
}

// KeyMatchesCategory 校验对象键是否属于指定目录前缀。
func KeyMatchesCategory(key string, cat Category) bool {
	key = strings.TrimLeft(strings.TrimSpace(key), "/")
	prefix := string(cat) + "/"
	return strings.HasPrefix(key, prefix)
}
