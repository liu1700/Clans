package utils

import (
	"Clans/server/log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// handle unix signals
func SigHandler(wg *sync.WaitGroup, shuttingDownChan chan struct{}) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM)

	for {
		msg := <-ch
		switch msg {
		case syscall.SIGTERM: // 关闭agent
			log.Logger().Info("sigterm received")
			log.Logger().Info("waiting for service close, please wait...")
			close(shuttingDownChan)
			log.Logger().Info("service shutdown.")
			wg.Done()
		}
	}
}

func CheckError(err error) {
	if err != nil {
		panic(err)
		os.Exit(-1)
	}
}
