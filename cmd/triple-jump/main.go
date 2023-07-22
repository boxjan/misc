package main

import (
	"context"
	"github.com/boxjan/misc/commom/cmd"
	"github.com/boxjan/misc/commom/signal"
	"github.com/huin/goupnp"
	"github.com/kardianos/service"
	"github.com/spf13/cobra"
	"github.com/valyala/bytebufferpool"
	"gopkg.in/yaml.v3"
	"k8s.io/klog/v2"
	"os"
	"runtime/debug"
	"time"
)

var (
	rootCmd        = cmd.QuickCobraRun("triple-jump", run)
	configPath     = &cmd.ConfigPath
	generateConfig = &cmd.GenerateConfig
	conf           = Conf{}

	// will give sometime for shutdown
	mainCtx context.Context
	// will cancel after sig catch
	subCtx context.Context
)

func init() {
	rootCmd.Short = "use in a router run with miniupnpd, get pub ip from upstream router, and sync upnp request to it"
}

func main() {
	if e := recover(); e != nil {
		trace := debug.Stack()
		klog.Errorf("catch panic: %v\nstack: %s", e, trace)
		return
	}

	var mainCtxCancel, subCtxCancel context.CancelFunc
	mainCtx, mainCtxCancel = context.WithCancel(context.Background())
	subCtx, subCtxCancel = context.WithCancel(mainCtx)

	signal.TrapSignals()

	signal.RegisterShutdownFunc(func() {
		subCtxCancel()
		// main ctx will cancel after all shutdown func executed
		time.Sleep(5 * time.Second)
		mainCtxCancel()
	})

	rootCmd.PreRunE = preRunE

	if err := rootCmd.Execute(); err != nil {
		klog.Error(err)
	}
}

func run(cmd *cobra.Command, args []string) {
	if generateConfig != nil && *generateConfig {
		b := bytebufferpool.Get()
		defer bytebufferpool.Put(b)
		err := yaml.NewEncoder(b).Encode(defaultConfig)
		if err != nil {
			klog.Fatalf("encode default config failed, err: %+v", err)
		}
		err = os.WriteFile(*configPath, b.Bytes(), 0644)
		if err != nil {
			klog.Fatalf("write default config failed, err: %+v", err)
		}
		return
	}

	instance := tripleJump{
		conf:                     conf,
		miniupnpdLeaseModifyTime: time.Time{},
	}
	var err error

	err = instance.prepareInstance()
	if err != nil {
		klog.Fatalf("prepare instance failed, err: %+v", err)
	}

	if instance.extInterfaceName != "" {
		klog.Infof("set search interface to %s", instance.extInterfaceName)
		goupnp.SetSearchInterface(instance.extInterfaceName)
	}

	err = instance.updateRouterClient(subCtx)
	if err != nil {
		klog.Fatalf("pick router client failed, err: %+v", err)
	}

	go instance.checkMiniupnpdLeaseFile(subCtx)
	go instance.checkUpstreamRouter(subCtx)

	<-mainCtx.Done()
}

func preRunE(cmd *cobra.Command, args []string) error {
	if d, err := os.ReadFile(*configPath); err != nil {
		klog.Warningf("can open load conf, err %+v, will use default config", err)
		conf = defaultConfig
		return nil
	} else if err := yaml.Unmarshal(d, &conf); err != nil {
		klog.Warningf("can not load conf, err %+v, will use default config", err)
		conf = defaultConfig
		return nil
	}
	return nil
}

type fakeHandle struct{}

func (p *fakeHandle) Start(_ service.Service) error { return nil }
func (p *fakeHandle) Stop(_ service.Service) error  { return nil }

func newServiceHandle(serviceName string) (service.Service, error) {
	return service.New(&fakeHandle{}, &service.Config{Name: serviceName})
}
