package setting

import (
	"context"
	"errors"
	"testing"
)

// fakeRepo is an in-memory stand-in for the settings store. It records the
// writes it receives so tests can assert whether the service reached the
// repository at all.
type fakeRepo struct {
	values map[string]string
	sets   []kv
	seeds  []kv
	getErr error
}

type kv struct{ key, value string }

func newFakeRepo(values map[string]string) *fakeRepo {
	if values == nil {
		values = map[string]string{}
	}
	return &fakeRepo{values: values}
}

func (f *fakeRepo) Get(_ context.Context, key string) (string, error) {
	if f.getErr != nil {
		return "", f.getErr
	}
	return f.values[key], nil
}

func (f *fakeRepo) GetAll(_ context.Context) (map[string]string, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	out := make(map[string]string, len(f.values))
	for k, v := range f.values {
		out[k] = v
	}
	return out, nil
}

func (f *fakeRepo) Set(_ context.Context, key, value string) error {
	f.sets = append(f.sets, kv{key, value})
	f.values[key] = value
	return nil
}

func (f *fakeRepo) Seed(_ context.Context, key, value string) error {
	f.seeds = append(f.seeds, kv{key, value})
	return nil
}

func TestGetAllMasksSensitiveValues(t *testing.T) {
	repo := newFakeRepo(map[string]string{
		"integration.monkeyapi_key": "sk-real-secret",
		"ui.theme":                  "dark",
	})
	svc := NewService(repo)

	all, err := svc.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}

	if got := all["integration.monkeyapi_key"]; got != masked {
		t.Errorf("sensitive key: got %q, want %q", got, masked)
	}
	if got := all["ui.theme"]; got != "dark" {
		t.Errorf("plain key: got %q, want %q", got, "dark")
	}
}

func TestGetAllLeavesEmptySensitiveValuesUnmasked(t *testing.T) {
	repo := newFakeRepo(map[string]string{"integration.paypinga_key": ""})
	svc := NewService(repo)

	all, err := svc.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}

	// An unset key must stay empty so the UI can tell "not configured"
	// apart from "configured but hidden".
	if got := all["integration.paypinga_key"]; got != "" {
		t.Errorf("unset sensitive key: got %q, want empty", got)
	}
}

func TestGetAllPropagatesRepoError(t *testing.T) {
	repo := newFakeRepo(nil)
	repo.getErr = errors.New("db down")
	svc := NewService(repo)

	if _, err := svc.GetAll(context.Background()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetReturnsRawSensitiveValue(t *testing.T) {
	repo := newFakeRepo(map[string]string{"integration.claude_api_key": "sk-real-secret"})
	svc := NewService(repo)

	// Get is the internal accessor used to build API clients, so it must
	// return the real credential. Only GetAll masks.
	got, err := svc.Get(context.Background(), "integration.claude_api_key")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != "sk-real-secret" {
		t.Errorf("got %q, want the raw value", got)
	}
}

func TestSetIgnoresMaskedSentinelForSensitiveKeys(t *testing.T) {
	repo := newFakeRepo(map[string]string{"integration.openai_api_key": "sk-real-secret"})
	svc := NewService(repo)

	if err := svc.Set(context.Background(), "integration.openai_api_key", masked); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Saving the settings form unchanged must not overwrite the stored
	// credential with the mask the UI displayed.
	if len(repo.sets) != 0 {
		t.Fatalf("expected no write, got %v", repo.sets)
	}
	if repo.values["integration.openai_api_key"] != "sk-real-secret" {
		t.Error("stored credential was overwritten by the mask")
	}
}

func TestSetWritesRealSensitiveValue(t *testing.T) {
	repo := newFakeRepo(map[string]string{"integration.openai_api_key": "sk-old"})
	svc := NewService(repo)

	if err := svc.Set(context.Background(), "integration.openai_api_key", "sk-new"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	if len(repo.sets) != 1 || repo.sets[0].value != "sk-new" {
		t.Fatalf("expected one write of the new value, got %v", repo.sets)
	}
}

func TestSetWritesMaskSentinelForNonSensitiveKeys(t *testing.T) {
	repo := newFakeRepo(nil)
	svc := NewService(repo)

	// The sentinel is only special for sensitive keys; elsewhere it is
	// just an ordinary string.
	if err := svc.Set(context.Background(), "ui.theme", masked); err != nil {
		t.Fatalf("Set: %v", err)
	}

	if len(repo.sets) != 1 || repo.sets[0].value != masked {
		t.Fatalf("expected the value to be written through, got %v", repo.sets)
	}
}
