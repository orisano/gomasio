package internal

import (
	"log"
	"sync"
)

var logMu sync.Mutex

func Log(v ...interface{}) {
	logMu.Lock()
	defer logMu.Unlock()
	log.Print(v)
}
