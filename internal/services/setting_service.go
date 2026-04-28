package services

import (
	"ai-later-nav/internal/database/repository"
	"sync"
)

var DefaultSettings = map[string]string{
	"copyright": "AI导航 © 2024",
	"site_name": "AI Later",
}

var (
	settingInstance *SettingService
	settingOnce    sync.Once
)

type SettingService struct {
	repo  *repository.SettingRepository
	mu    sync.RWMutex
	cache map[string]string
}

func GetSettingService() *SettingService {
	settingOnce.Do(func() {
		settingInstance = &SettingService{
			repo:  repository.NewSettingRepository(),
			cache: make(map[string]string),
		}
	})
	return settingInstance
}

func (s *SettingService) LoadCache() error {
	all, err := s.repo.GetAll()
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.cache = all
	s.mu.Unlock()
	return nil
}

func (s *SettingService) GetSetting(key string) string {
	s.mu.RLock()
	v := s.cache[key]
	s.mu.RUnlock()
	if v != "" {
		return v
	}
	if def, ok := DefaultSettings[key]; ok {
		return def
	}
	return ""
}

func (s *SettingService) GetAllSettings() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]string, len(DefaultSettings))
	for k, v := range DefaultSettings {
		result[k] = v
	}
	for k, v := range s.cache {
		result[k] = v
	}
	return result
}

func (s *SettingService) UpdateSetting(key, value string) error {
	if err := s.repo.Set(key, value); err != nil {
		return err
	}
	s.mu.Lock()
	s.cache[key] = value
	s.mu.Unlock()
	return nil
}

func (s *SettingService) UpdateMultiple(settings map[string]string) error {
	if err := s.repo.SetMultiple(settings); err != nil {
		return err
	}
	s.mu.Lock()
	for k, v := range settings {
		s.cache[k] = v
	}
	s.mu.Unlock()
	return nil
}

func (s *SettingService) SeedDefaults() error {
	if err := s.repo.SeedDefaults(DefaultSettings); err != nil {
		return err
	}
	return s.LoadCache()
}
