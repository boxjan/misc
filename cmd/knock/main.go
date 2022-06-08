package main

import (
	"context"
	"fmt"
	"github.com/boxjan/misc/commom/address"
	"github.com/boxjan/misc/commom/cidrset"
	"github.com/boxjan/misc/commom/cmd"
	. "github.com/boxjan/misc/commom/fiber"
	"github.com/boxjan/misc/pkg/knock/ent"
	"github.com/boxjan/misc/pkg/knock/ent/wireguardclient"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"k8s.io/klog/v2"
	"net"
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
	b, _ := yaml.Marshal(defaultConfig)
	fmt.Println(b)
}

func main() {
	if e := recover(); e != nil {
		trace := debug.Stack()
		klog.Errorf("catch panic: %v\nstack: %s", e, trace)
		return
	}

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
	var err error

	if _, cidr, err = net.ParseCIDR(conf.Wireguard.AllocCidr); err == nil {
		cidrSet, err = cidrset.NewCIDRSet(cidr, 30)
	}
}

func preRunE() error {
	// client database
	var err error

	// new alloc cidr set
	var cidr *net.IPNet
	if _, cidr, err = net.ParseCIDR(conf.Wireguard.AllocCidr); err == nil {
		cidrSet, err = cidrset.NewCIDRSet(cidr, 30)
	}
	if err != nil {
		return fmt.Errorf("new cidr set failed with err %v", err)
	}

	// mark in use alloc cidr in cidr set
	wc, err := dbCli.WireguardClient.Query().Where(wireguardclient.Expired(false)).All(context.Background())
	if err != nil {
		return fmt.Errorf("get alloced cidr list failed with err: %s", err)
	}
	for _, w := range wc {
		_, c, err := net.ParseCIDR(w.AllocCidr)
		if err == nil {
			_ = cidrSet.Occupy(c)
		}
	}

	for _, i := range conf.Wireguard.AllowedIps {
		_, c, err := net.ParseCIDR(i)
		if err != nil {
			return fmt.Errorf("parse all")
		}
		conf.Wireguard.allowedIps = append(conf.Wireguard.allowedIps, *c)
	}

	return nil
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
