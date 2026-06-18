package setting

import (
	"context"

	"github.com/husari/hube/internal/infrastructure/sqlite"
)

const masked = "••••••••"

var sensitiveKeys = map[string]bool{
	"integration.monkeyapi_key": true,
	"integration.paypinga_key":  true,
}

type Service struct{ repo *sqlite.SettingRepo }

func NewService(repo *sqlite.SettingRepo) *Service { return &Service{repo: repo} }

func (s *Service) Get(ctx context.Context, key string) (string, error) {
	return s.repo.Get(ctx, key)
}

func (s *Service) GetAll(ctx context.Context) (map[string]string, error) {
	all, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	for k := range sensitiveKeys {
		if v := all[k]; v != "" {
			all[k] = masked
		}
	}
	return all, nil
}

// Set persists a key. If the value equals the masked sentinel and the key is sensitive, it is a no-op.
func (s *Service) Set(ctx context.Context, key, value string) error {
	if sensitiveKeys[key] && value == masked {
		return nil
	}
	return s.repo.Set(ctx, key, value)
}

func (s *Service) Seed(ctx context.Context, key, value string) error {
	return s.repo.Seed(ctx, key, value)
}
