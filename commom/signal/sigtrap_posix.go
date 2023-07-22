//go:build !windows && !plan9 && !nacl && !js
// +build !windows,!plan9,!nacl,!js

package signal

import (
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

func trapSignalsPosix() {
	go func() {
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGQUIT)

		for i := 0; true; i++ {
			sig := <-shutdown
			var sn string

			switch sig {
			case syscall.SIGTERM:
				sn = "SIGTERM"
			case syscall.SIGQUIT:
				sn = "SIGQUIT"
			}

			zap.L().Info("caught " + sn + ", will shutting down")
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
