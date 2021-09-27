package wait

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// ForCtrlC https://jjasonclark.com/waiting_for_ctrl_c_in_golang/
func ForCtrlC() {
	var endWaiter sync.WaitGroup
	endWaiter.Add(1)
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-signalChannel
		endWaiter.Done()
	}()
	endWaiter.Wait()
}
