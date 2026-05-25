package media

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"xlangai/server/config"
	"xlangai/server/internal/objectstore"
	"xlangai/server/internal/repository"
	"xlangai/server/internal/settings"
	"xlangai/server/internal/storage"
)

var (
	ErrClientOnlyStorage = errors.New("storage policy is client only")
	ErrDirectUploadNA    = errors.New("direct upload not available for current storage configuration")
	ErrInvalidObjectURL  = errors.New("invalid object url for this media scope")
)

// StorageScope 对应 sys_system_settings 中的媒体策略 key。
type StorageScope int

const (
	ScopeUserRecording StorageScope = iota
	ScopeAssistantTTS
	ScopeAvatar
)

// Service 按系统变量中的媒体策略与 sys_object_storage_configs 上传；策略与厂商表解耦。
type Service struct {
	repo     *repository.ObjectStorageConfigRepo
	cfg      *config.Config
	settings *settings.Service
}

func NewService(repo *repository.ObjectStorageConfigRepo, cfg *config.Config, sys *settings.Service) *Service {
	return &Service{repo: repo, cfg: cfg, settings: sys}
}

func (s *Service) policyKey(scope StorageScope) string {
	switch scope {
	case ScopeUserRecording:
		return settings.MediaUserRecordingStorage
	case ScopeAssistantTTS:
		return settings.MediaAssistantTtsStorage
	case ScopeAvatar:
		return settings.MediaAvatarStorage
	default:
		return settings.MediaUserRecordingStorage
	}
}

func (s *Service) categoryForScope(scope StorageScope) objectstore.Category {
	switch scope {
	case ScopeAvatar:
		return objectstore.CategoryAvatar
	case ScopeAssistantTTS:
		return objectstore.CategoryAIAudio
	default:
		return objectstore.CategoryUserAudio
	}
}

func (s *Service) localDirs() objectstore.LocalDirs {
	av := strings.TrimSpace(s.cfg.AvatarDir)
	if av == "" {
		av = "./storage/avatars"
	}
	ad := strings.TrimSpace(s.cfg.AudioDir)
	if ad == "" {
		ad = "./storage/audio"
	}
	return objectstore.LocalDirs{AvatarDir: av, AudioDir: ad}
}

func (s *Service) active(ctx context.Context) (*objectstore.RuntimeConfig, error) {
	c, err := s.repo.GetActive(ctx)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, nil
	}
	return objectstore.FromRepo(c), nil
}

func (s *Service) storageMode(ctx context.Context, scope StorageScope) string {
	mode := "server"
	if s.settings != nil {
		mode = strings.ToLower(strings.TrimSpace(s.settings.String(ctx, s.policyKey(scope))))
	}
	if mode == "" {
		if def, ok := settings.Defaults[s.policyKey(scope)]; ok {
			mode = def
		}
	}
	return mode
}

func (s *Service) upload(ctx context.Context, scope StorageScope, in objectstore.UploadInput) (*objectstore.UploadResult, error) {
	mode := s.storageMode(ctx, scope)
	switch mode {
	case "client":
		return nil, ErrClientOnlyStorage
	case "cloud":
		rc, err := s.active(ctx)
		if err != nil {
			return nil, err
		}
		if rc == nil {
			return nil, objectstore.ErrNotConfigured
		}
		return objectstore.Upload(ctx, rc, s.localDirs(), in)
	case "server":
		return objectstore.Upload(ctx, nil, s.localDirs(), in)
	default:
		return objectstore.Upload(ctx, nil, s.localDirs(), in)
	}
}

// SupportsDirectUpload 云存储且 provider 支持预签名 PUT 时为 true。
func (s *Service) SupportsDirectUpload(ctx context.Context, scope StorageScope) (bool, error) {
	if s.storageMode(ctx, scope) != "cloud" {
		return false, nil
	}
	rc, err := s.active(ctx)
	if err != nil {
		return false, err
	}
	if rc == nil {
		return false, nil
	}
	p := strings.ToLower(strings.TrimSpace(rc.Provider))
	return p == objectstore.ProviderCloudflareR2 || p == objectstore.ProviderAliyunOSS, nil
}

// PresignUpload 签发客户端直传凭证。
func (s *Service) PresignUpload(ctx context.Context, scope StorageScope, ext, contentType string) (*objectstore.PresignResult, error) {
	if s.storageMode(ctx, scope) == "client" {
		return nil, ErrClientOnlyStorage
	}
	reason, rc, err := s.presignDiagnostics(ctx, scope)
	if err != nil {
		return nil, err
	}
	if reason != "" {
		return nil, &PresignBlocked{Reason: reason}
	}
	return objectstore.PresignPut(ctx, rc, objectstore.PresignInput{
		Category:    s.categoryForScope(scope),
		Ext:         ext,
		ContentType: contentType,
		TTL:         15 * time.Minute,
	})
}

// ValidateObjectURL 校验 URL 属于当前 active 配置的公网域名与目录前缀。
func (s *Service) ValidateObjectURL(ctx context.Context, scope StorageScope, objectURL string) error {
	if s.storageMode(ctx, scope) != "cloud" {
		return ErrInvalidObjectURL
	}
	rc, err := s.active(ctx)
	if err != nil {
		return err
	}
	if rc == nil {
		return objectstore.ErrNotConfigured
	}
	key := objectstore.KeyFromPublicURL(rc.PublicBaseURL, objectURL)
	if key == "" || !objectstore.KeyMatchesCategory(key, s.categoryForScope(scope)) {
		return ErrInvalidObjectURL
	}
	return nil
}

// DownloadObjectBytes 按公网 URL 从云存储拉取对象（用于已直传后的 STT 等）。
func (s *Service) DownloadObjectBytes(ctx context.Context, objectURL string) ([]byte, error) {
	rc, err := s.active(ctx)
	if err != nil {
		return nil, err
	}
	if rc == nil {
		return nil, objectstore.ErrNotConfigured
	}
	return objectstore.DownloadByPublicURL(ctx, rc, objectURL)
}

func (s *Service) SaveAvatar(ctx context.Context, data []byte, ext, contentType string) (*objectstore.UploadResult, error) {
	return s.upload(ctx, ScopeAvatar, objectstore.UploadInput{
		Category: objectstore.CategoryAvatar, Data: data, Ext: ext, ContentType: contentType,
	})
}

func (s *Service) SaveUserRecording(ctx context.Context, data []byte, ext, contentType string) (*objectstore.UploadResult, error) {
	return s.upload(ctx, ScopeUserRecording, objectstore.UploadInput{
		Category: objectstore.CategoryUserAudio, Data: data, Ext: ext, ContentType: contentType,
	})
}

func (s *Service) SaveAssistantTTS(ctx context.Context, data []byte, ext, contentType string) (*objectstore.UploadResult, error) {
	return s.upload(ctx, ScopeAssistantTTS, objectstore.UploadInput{
		Category: objectstore.CategoryAIAudio, Data: data, Ext: ext, ContentType: contentType,
	})
}

func (s *Service) ReadAvatar(name string) ([]byte, error) {
	return storage.ReadAudio(s.localDirs().AvatarDir, name)
}

func (s *Service) ReadAudio(name string) ([]byte, error) {
	dirs := s.localDirs()
	data, err := storage.ReadAudio(dirs.AudioDir, name)
	if err == nil {
		return data, nil
	}
	bundledDir := strings.TrimSpace(os.Getenv("BUNDLED_AUDIO_DIR"))
	if bundledDir == "" {
		bundledDir = "/app/bootstrap-storage/audio"
	}
	if bundledDir != "" && bundledDir != dirs.AudioDir {
		if fallback, fallbackErr := storage.ReadAudio(bundledDir, name); fallbackErr == nil {
			return fallback, nil
		}
	}
	return nil, err
}

// Capabilities 返回各媒体域的存储模式与是否支持直传（供客户端决策）。
func (s *Service) Capabilities(ctx context.Context) (map[string]interface{}, error) {
	type scopeCap struct {
		Storage      string `json:"storage"`
		DirectUpload bool   `json:"direct_upload"`
	}
	out := map[string]interface{}{
		"provider": "",
		"scopes":   map[string]scopeCap{},
	}
	rc, err := s.active(ctx)
	if err != nil {
		return nil, err
	}
	if rc != nil {
		out["provider"] = strings.TrimSpace(rc.Provider)
	}
	scopes := map[string]scopeCap{}
	for name, scope := range map[string]StorageScope{
		"avatar":         ScopeAvatar,
		"user_recording": ScopeUserRecording,
		"assistant_tts":  ScopeAssistantTTS,
	} {
		mode := s.storageMode(ctx, scope)
		direct, derr := s.SupportsDirectUpload(ctx, scope)
		if derr != nil {
			return nil, fmt.Errorf("capabilities %s: %w", name, derr)
		}
		scopes[name] = scopeCap{Storage: mode, DirectUpload: direct && mode == "cloud"}
	}
	out["scopes"] = scopes
	return out, nil
}
