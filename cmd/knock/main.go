package main

import (
	"context"
	"github.com/boxjan/misc/commom/address"
	"github.com/boxjan/misc/commom/cidrset"
	"github.com/boxjan/misc/commom/cmd"
	. "github.com/boxjan/misc/commom/fiber"
	"github.com/boxjan/misc/pkg/knock/ent"
	"github.com/boxjan/misc/pkg/knock/ent/wireguardclient"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"golang.zx2c4.com/wireguard/wgctrl"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"k8s.io/klog/v2"
	"net"
	"os"
	"runtime/debug"
	"strings"
)

var (
	rootCmd        = cmd.QuickCobraRun("knock", run)
	configPath     = &cmd.ConfigPath
	generateConfig = &cmd.GenerateConfig
	conf           = Config{}

	dbCli   = &ent.Client{}
	cidrSet *cidrset.CidrSet
)

func init() {
}

func main() {
	defer func() {
		if e := recover(); e != nil {
			trace := debug.Stack()
			klog.Errorf("catch panic: %v\nstack: %s", e, trace)
			return
		}
	}()

	if err := rootCmd.Execute(); err != nil {
		klog.Error(err)
	}
}

func writeConfig() {
	if len(*configPath) != 0 {
		b, err := yaml.Marshal(conf)
		if err != nil {
			klog.Errorf("generate config file %s failed with err: %s", *configPath, err)
		}
		if err = ioutil.WriteFile(*configPath, b, 0644); err != nil {
			klog.Errorf("generate config file %s failed with err: %s", *configPath, err)
		}
	}
}

func loadConfig() {
	if d, err := ioutil.ReadFile(*configPath); err != nil {
		klog.Warningf("can not load conf, will use default config")
		conf = defaultConfig
	} else if err := yaml.Unmarshal(d, &conf); err != nil {
		klog.Warningf("can not load conf, will use default config")
		conf = defaultConfig
	}

	// get connect ip
	if conf.Wireguard.LocalIp == "" {
		pubIp := address.GetMyPubIpv4()
		if pubIp == "" {
			klog.Fatal("need a ip write in client conf")
		}
		conf.Wireguard.LocalIp = strings.Trim(pubIp, "\n")
	}

	for _, i := range conf.Wireguard.AllowedIps {
		_, c, err := net.ParseCIDR(i)
		if err != nil {
			klog.Fatalf("parse cidr: %s failed with err: %s", i, err)
		}
		conf.Wireguard.allowedIps = append(conf.Wireguard.allowedIps, *c)
	}
}

func initDatabase() {
	var err error
	dbCli, err = ent.Open(conf.Database.Type, conf.Database.Dsn)
	if err != nil {
		klog.Fatalf("client database failed with err: %s", err)
	}
	if err = dbCli.Schema.Create(context.Background()); err != nil {
		klog.Fatalf("failed creating schema resources: %v", err)
	}
}

func initAllocAddrPool() {
	var cidr *net.IPNet
	var wgCli *wgctrl.Client
	var err error

	if _, cidr, err = net.ParseCIDR(conf.Wireguard.AllocCidr); err == nil {
		cidrSet, err = cidrset.NewCIDRSet(cidr, 30)
	}

	for _, eac := range conf.Wireguard.ExcludeAllocCidr {
		if _, i, err := net.ParseCIDR(eac); err == nil {
			if err := cidrSet.Occupy(i); err != nil {
				klog.Warningf("occupy cidr set meet err: %v", err)
			}
		} else {
			klog.Fatalf("parse ExcludeAllocCidr: %s failed with err: %v", eac, err)
		}
	}

	// mark in use alloc cidr in cidr set
	wgcs, err := dbCli.WireguardClient.Query().Where(wireguardclient.Expired(false)).All(context.Background())
	if err != nil {
		klog.Fatalf("get active wireguard link from database failed with err: %s", err)
	}

	wgCli, err = wgctrl.New()
	if err != nil {
		klog.Fatalf("init wireguard ctrl failed with err: %v", err)
	}

	for _, wgc := range wgcs {
		if _, err := wgCli.Device(wgc.NetifName); err != nil {
			if err == os.ErrNotExist {
				klog.Infof("device %s not exists any more, will update database", wgc.NetifName)
				if _, err := wgc.Update().SetExpired(true).Save(context.Background()); err != nil {
					klog.Warningf("save %s failed with err: %s", wgc.NetifName, err)
				}
			} else {
				klog.Warningf("device %s exists in database but sync info failed with err: %s", wgc.NetifName, err)
			}
			continue
		}

		if _, c, err := net.ParseCIDR(wgc.AllocCidr); err == nil {
			if err := cidrSet.Occupy(c); err != nil {
				klog.Warningf("occupy cidr set meet err: %v", err)
			}
		} else {
			klog.Fatalf("parse exists device %s, cidr: %s failed with err: %v", c, err)
		}
	}
}

func run(cmd *cobra.Command, args []string) {
	if *generateConfig {
		writeConfig()
	}

	loadConfig()
	initDatabase()
	initAllocAddrPool()

	app := DefaultFiber()

	apiV1 := app.Group("/api/v1").Name("api-v1/")
	apiV1.Get("knock", NewWireguardHandle).Name("knock")

	go wireguardBackground()

	klog.Fatal(app.Listen(conf.Addr))
}
