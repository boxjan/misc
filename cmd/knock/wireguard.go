package main

import (
	"context"
	"fmt"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
)

type WgConfig struct {
	PrivateKey wgtypes.Key
	ListenPort *int
	Address    string
	Peers      []wgtypes.PeerConfig
}

func wireguardBackground() {

}

func NewWireguard(identify string) (ClientConf *WgConfig, err error) {

	var wgc = dbCli.WireguardClient.Create().SetIdentify(identify)

	var tunnelIps *net.IPNet
	tunnelIps, err = cidrSet.AllocateNext()

	ServerConf := &WgConfig{}
	ClientConf = &WgConfig{}

	ServerConf.PrivateKey, err = wgtypes.GeneratePrivateKey()
	if err != nil {
		err = fmt.Errorf("gen private key failed with err")
		return
	}
	ClientConf.PrivateKey, err = wgtypes.GeneratePrivateKey()
	if err != nil {
		err = fmt.Errorf("gen private key failed with err")
		return
	}

	ServerConfPeer := wgtypes.PeerConfig{}
	ServerConfPeer.PublicKey = ClientConf.PrivateKey.PublicKey()

	ClientConfPeer := wgtypes.PeerConfig{}
	ClientConfPeer.PublicKey = ServerConf.PrivateKey.PublicKey()

	clientIp := tunnelIps.IP
	if len(clientIp) == net.IPv4len {
		clientIp[3] = clientIp[3] + 2
	} else {
		clientIp[15] = clientIp[15] + 2
	}

	serverIp := tunnelIps.IP
	if len(serverIp) == net.IPv4len {
		serverIp[3] = serverIp[3] + 2
	} else {
		serverIp[15] = serverIp[15] + 2
	}

	ServerConf.Address = serverIp.String() + "/" + tunnelIps.Mask.String()
	ClientConf.Address = clientIp.String() + "/" + tunnelIps.Mask.String()

	l, err := net.Listen("udp", "[::]:0")
	if err != nil {
		err = fmt.Errorf("can not found a ")
	}
	port := l.Addr().(*net.UDPAddr).Port

	ClientConfPeer.Endpoint = &net.UDPAddr{IP: net.ParseIP(conf.Wireguard.LocalIp), Port: port}
	ServerConf.ListenPort = &port

	ClientConfPeer.AllowedIPs = conf.Wireguard.allowedIps

	wgc.Save(context.Background())
	return
}

//func DestroyWireguard(ServerConf wgtypes.Config) error {
//	ws, err := dbCli.WireguardClient.Query().
//		Where(wireguardclient.ServerPrivateKey(ServerConf.PrivateKey.String())).All(context.Background())
//
//}
