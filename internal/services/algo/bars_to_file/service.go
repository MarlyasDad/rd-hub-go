package barstofile

import (
	"github.com/MarlyasDad/rd-hub-go/pkg/alor"
	"os"
	"path/filepath"
)

type Service struct {
	Name    string
	File    *os.File
	History *alor.BarQueue
}

func New(filename string) *Service {
	return &Service{Name: filename}
}

func (s *Service) SetHistory(history *alor.BarQueue) {
	s.History = history
}

func (s *Service) Init() error {
	file, err := os.Create(filepath.Join("data", s.Name))
	if err != nil {
		return err
	}
	s.File = file
	return nil
}

func (s *Service) DeInit() error {
	return s.File.Close()
}
