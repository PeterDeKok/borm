package borm

import "peterdekok.nl/gotools/logger"

var log logger.Logger

func init() {
    log = logger.New("borm")
}