package barstofile

import (
	"encoding/json"
	"fmt"
	"github.com/MarlyasDad/rd-hub-go/pkg/alor"
	"log"
)

func (s *Service) NewBar(data alor.BarsSlimData) error {
	log.Println("NewBar", data.Time)

	bar, err := s.History.GetLastFinalizedBar()
	if err != nil {
		return err
	}

	barBytes, err := json.Marshal(bar)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(s.File, barBytes)
	if err != nil {
		return err
	}

	return nil
}
