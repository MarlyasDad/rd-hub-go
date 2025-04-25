package barstofile

import (
	"github.com/MarlyasDad/rd-hub-go/pkg/alor"
	"log"
)

func (s *Service) NewAllTrades(data alor.AllTradesSlimData) error {
	log.Println("service new all trades", data)
	return nil
}
