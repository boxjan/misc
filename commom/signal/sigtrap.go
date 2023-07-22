package signal

import (
	"go.uber.org/zap"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func TrapSignals() {
	trapSignalsCrossPlatform()
	trapSignalsPosix()
}

var shutdownFunc []func()
var callOnceShutdownFunc sync.Once

func RegisterShutdownFunc(f func()) {
	shutdownFunc = append(shutdownFunc, f)
}

// trapSignalsCrossPlatform captures SIGINT, which triggers forceful
// shutdown that executes shutdown callbacks first. A second interrupt
// signal will exit the process immediately.
func trapSignalsCrossPlatform() {
	go func() {
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

		for i := 0; true; i++ {
			<-shutdown
			zap.L().Info("caught SIGINT, shutting down")

			callOnceShutdownFunc.Do(
				func() {
					for _, f := range shutdownFunc {
						go f()
					}
					zap.L().Info("call all shutdown func complete")
				},
			)
		}
	}()
}
