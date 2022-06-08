package main

import "net"

type Config struct {
	Addr      string
	Database  Database  `yaml:"database"`
	Wireguard Wireguard `yaml:"wireguard"`
	Http      Http      `yaml:"http"`
}

type Database struct {
	Type string `yaml:"type"`
	Dsn  string `yaml:"dsn"`
}

type Wireguard struct {
	AllocCidr        string   `yaml:"alloc_cidr"`
	ExcludeAllocCidr []string `yaml:"exclude_alloc_cidr"`
	LocalIp          string   `yaml:"local_ip"`
	AllowedIps       []string `yaml:"allowed_ips"`
	allowedIps       []net.IPNet
}

type Http struct {
	IdentityHeader string `yaml:"identity_header"`
}

var defaultConfig = Config{
	Addr: "[::]:8000",
	Database: Database{
		Type: "sqlite3",
		Dsn:  "file:ent?mode=memory&_fk=1",
	},
	Wireguard: Wireguard{
		AllocCidr:        "10.100.0.0/24",
		ExcludeAllocCidr: []string{"10.100.0.0/26"},
		LocalIp:          "",
		AllowedIps:       []string{"10.100.0.0/16", "10.105.0.0/16"},
	},
	Http: Http{
		IdentityHeader: "",
	},
}
