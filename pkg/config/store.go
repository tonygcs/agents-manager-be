package config

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// ErrInUse is returned when a resource cannot be deleted because it is referenced by another.
type ErrInUse struct {
	Resource   string
	ReferencedBy string
}

func (e *ErrInUse) Error() string {
	return fmt.Sprintf("%q is used by %q", e.Resource, e.ReferencedBy)
}

// Store provides thread-safe access to the configuration and persists changes to disk.
type Store struct {
	mu   sync.RWMutex
	path string
	cfg  Config
}

func NewStore(path string) (*Store, error) {
	cfg, err := Load(path)
	if err != nil {
		return nil, err
	}
	return &Store{path: path, cfg: cfg}, nil
}

func (s *Store) WorkerdAddr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg.Workerd.Addr
}

func (s *Store) ServerAddr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg.Server.Addr
}

func (s *Store) Secrets() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]string, len(s.cfg.Secrets))
	for k, v := range s.cfg.Secrets {
		out[k] = v
	}
	return out
}

func (s *Store) Secret(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.cfg.Secrets[key]
	return v, ok
}

func (s *Store) SetSecret(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cfg.Secrets == nil {
		s.cfg.Secrets = make(map[string]string)
	}
	s.cfg.Secrets[key] = value
	return s.save()
}

func (s *Store) DeleteSecret(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.cfg.Secrets[key]; !ok {
		return fmt.Errorf("secret %q not found", key)
	}
	for workerName, wc := range s.cfg.Workers {
		for _, secret := range wc.Secrets {
			if secret == key {
				return &ErrInUse{Resource: key, ReferencedBy: workerName}
			}
		}
	}
	delete(s.cfg.Secrets, key)
	return s.save()
}

func (s *Store) Workers() map[string]WorkerConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]WorkerConfig, len(s.cfg.Workers))
	for k, v := range s.cfg.Workers {
		out[k] = v
	}
	return out
}

func (s *Store) Worker(name string) (WorkerConfig, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w, ok := s.cfg.Workers[name]
	return w, ok
}

func (s *Store) SetWorker(name string, wc WorkerConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cfg.Workers == nil {
		s.cfg.Workers = make(map[string]WorkerConfig)
	}
	s.cfg.Workers[name] = wc
	return s.save()
}

func (s *Store) DeleteWorker(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.cfg.Workers[name]; !ok {
		return fmt.Errorf("worker %q not found", name)
	}
	delete(s.cfg.Workers, name)
	return s.save()
}

// save writes the current config to disk. Caller must hold mu.
func (s *Store) save() error {
	f, err := os.Create(s.path)
	if err != nil {
		return fmt.Errorf("config: save: %w", err)
	}
	defer f.Close()
	if err := yaml.NewEncoder(f).Encode(s.cfg); err != nil {
		return fmt.Errorf("config: save: %w", err)
	}
	return nil
}
