package media

import (
	"context"
	"fmt"

	"xlangai/server/internal/objectstore"
)

// PresignBlocked 预签名被拒绝（配置不满足），Reason 为中文说明。
type PresignBlocked struct {
	Reason string
}

func (e *PresignBlocked) Error() string {
	if e == nil || e.Reason == "" {
		return "presign blocked"
	}
	return e.Reason
}

// presignDiagnostics 返回当前 scope 下无法预签名的具体原因；空字符串表示可以预签名。
func (s *Service) presignDiagnostics(ctx context.Context, scope StorageScope) (string, *objectstore.RuntimeConfig, error) {
	mode := s.storageMode(ctx, scope)
	key := s.policyKey(scope)
	if mode == "client" {
		return fmt.Sprintf("系统设置 %s=%q：为 client 时不走云存储", key, mode), nil, nil
	}
	if mode != "cloud" {
		return fmt.Sprintf(
			"系统设置 %s=%q：须改为 cloud（仅配 R2 不够，还要在「系统设置」里把头像存储设为 cloud）",
			key, mode,
		), nil, nil
	}
	rc, err := s.active(ctx)
	if err != nil {
		return "", nil, err
	}
	if rc == nil {
		return "对象存储：没有 status=active 的配置，请在后台启用一条云存储配置", nil, nil
	}
	if !objectstore.SupportsDirectUploadProvider(rc.Provider) {
		return fmt.Sprintf(
			"对象存储：当前启用的 provider=%q 不支持客户端直传",
			rc.Provider,
		), rc, nil
	}
	return "", rc, nil
}
