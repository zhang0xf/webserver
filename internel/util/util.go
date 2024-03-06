package util

import (
	"os"
	"os/signal"
	"syscall"
)

func WaitForTerminate() {
	exitChan := make(chan struct{})
	signalChan := make(chan os.Signal, 1)
	go func() {
		<-signalChan
		close(exitChan)
	}()
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-exitChan
}
