package alor

import (
	"fmt"
	"sync"
)

type Storage struct {
	FlagStorage map[string]bool    `json:"flag_storage"`
	TextStorage map[string]string  `json:"text_storage"`
	IntStorage  map[string]float64 `json:"int_storage"`
	DecStorage  map[string]float64 `json:"dec_storage"`
	mu          sync.RWMutex
}

func newStorage() *Storage {
	return &Storage{
		FlagStorage: make(map[string]bool),
		TextStorage: make(map[string]string),
		IntStorage:  make(map[string]float64),
		DecStorage:  make(map[string]float64),
	}
}

func (s *Storage) SetFlag(name string, value bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.FlagStorage[name] = value
}

func (s *Storage) GetFlag(name string) (bool, error) {
	s.mu.RLock()
	defer s.mu.Unlock()

	flag, ok := s.FlagStorage[name]
	if !ok {
		return flag, fmt.Errorf("no flag found with name %s", name)
	}

	return flag, nil
}
