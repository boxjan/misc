package main

import "runtime"

type Conf struct {
	MiniupnpdLeasePath    string         `yaml:"miniupnpd-lease-path"`
	UpdateMiniupnpdConfig bool           `yaml:"update-miniupnpd-config"`
	MiniupnpdConfigPath   string         `yaml:"miniupnpd-config-path"`
	MiniupnpdServiceName  string         `yaml:"miniupnpd-service-name"`
	AutoRestartMiniupnpd  bool           `yaml:"auto-restart-miniupnpd"`
	ExtraForwards         []ExtraForward `yaml:"extra-forwards"`
}

type ExtraForward struct {
	// ip port external_port protocol
	Ip           string `yaml:"ip"`
	Port         uint16 `yaml:"port"`
	ExternalPort uint16 `yaml:"external-port"`
	Protocol     string `yaml:"protocol"`
}

var defaultConfig = Conf{
	MiniupnpdLeasePath:    "/var/lib/miniupnpd/upnp.leases",
	UpdateMiniupnpdConfig: true,
	MiniupnpdConfigPath:   "/etc/miniupnpd/miniupnpd.conf",
	MiniupnpdServiceName:  "miniupnpd",
	ExtraForwards:         nil,
}

func init() {
	switch runtime.GOOS {
	case "linux":
		conf.AutoRestartMiniupnpd = true
	default:
		conf.AutoRestartMiniupnpd = false
	}
}
