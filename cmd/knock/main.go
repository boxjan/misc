package main

import (
	"context"
	"fmt"
	"github.com/boxjan/misc/commom/address"
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
	rootCmd    = cmd.QuickCobraRun("knock", run)
	configPath = &cmd.ConfigPath
	conf       = Config{}

	dbCli   = &ent.Client{}
	cidrSet *address.CidrSet
)

func init() {

}

func main() {
	if e := recover(); e != nil {
		trace := debug.Stack()
		klog.Errorf("catch panic: %v\nstack: %s", e, trace)
		return
	}

	rootCmd.PreRunE = preRunE

	if err := rootCmd.Execute(); err != nil {
		klog.Error(err)
	}
}

func preRunE(cmd *cobra.Command, args []string) error {
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
			return fmt.Errorf("emm, need a ip write in client conf")
		}
		conf.Wireguard.LocalIp = strings.Trim(pubIp, "\n")
	}

	// client database
	var err error
	dbCli, err = ent.Open(conf.Database.Type, conf.Database.Dsn)
	if err != nil {
		return fmt.Errorf("client database failed with err: %s", err)
	}

	// finish migration database
	if err := dbCli.Schema.Create(context.Background()); err != nil {
		return fmt.Errorf("failed creating schema resources: %v", err)
	}

	// new alloc cidr set
	var cidr *net.IPNet
	if _, cidr, err = net.ParseCIDR(conf.Wireguard.AllocCidr); err == nil {
		cidrSet, err = address.NewCIDRSet(cidr, 30)
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
	app := DefaultFiber()

	apiV1 := app.Group("/api/v1").Name("api-v1/")
	apiV1.Get("knock", NewWireguardHandle).Name("knock")

	go wireguardBackground()

	klog.Fatal(app.Listen(conf.Addr))
}
