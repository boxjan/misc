package main

import "net"

type Config struct {
	Addr      string
	Database  Database  `yaml:"database"`
	Wireguard Wireguard `yaml:"wireguard"`
	Http      Http      `yaml:"http"`
}

type Database struct {
	Type string
	Dsn  string
}

type Wireguard struct {
	AllocCidr  string
	LocalIp    string
	AllowedIps []string
	allowedIps []net.IPNet
}

type Http struct {
	IdentityHeader string
}

var defaultConfig = Config{
	Addr: "[::]:8000",
	Database: Database{
		Type: "sqlite3",
		Dsn:  "file:ent?mode=memory&_fk=1",
	},
	Wireguard: Wireguard{
		AllocCidr:  "10.100.255.128/25",
		LocalIp:    "",
		AllowedIps: []string{"10.100.0.0/16", "10.105.0.0/16"},
	},
	Http: Http{
		IdentityHeader: "",
	},
}
