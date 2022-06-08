package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"io/ioutil"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type WgConfig struct {
	PrivateKey wgtypes.Key
	ListenPort *int
	Address    string
	Peers      []wgtypes.PeerConfig
}

func (cfg *WgConfig) Parse() []byte {
	w := &bytes.Buffer{}
	w.WriteString("[Interface]\n")

	fmt.Fprintf(w, "PrivateKey = %s\n", cfg.PrivateKey.String())
	if cfg.ListenPort != nil {
		fmt.Fprintf(w, "ListenPort = %d\n", *cfg.ListenPort)
	}
	fmt.Fprintf(w, "Address = %s\n", cfg.Address)

	w.WriteString("[Peer]\n")

	for _, p := range cfg.Peers {
		if p.Endpoint != nil {
			fmt.Fprintf(w, "Endpoint = %s\n", p.Endpoint.String())
		}
		fmt.Fprintf(w, "PublicKey = %s\n", p.PublicKey.String())
		fmt.Fprintf(w, "PersistentKeepalive = %d\n", 10)
		fmt.Fprintf(w, "AllowedIPs = %s\n", strings.Join(conf.Wireguard.AllowedIps, ", "))
	}
	return w.Bytes()
}

func wireguardBackground() {

}

func NewWireguard(identify string) (ClientConf *WgConfig, err error) {

	var wgc = dbCli.WireguardClient.Create().SetIdentify(identify)

	var tunnelIps *net.IPNet
	tunnelIps, err = cidrSet.AllocateNext()
	wgc.SetAllocCidr(tunnelIps.String())

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
	wgc.SetServerPrivateKey(hex.EncodeToString(ServerConf.PrivateKey[:]))
	wgc.SetClientPrivateKey(hex.EncodeToString(ClientConf.PrivateKey[:]))

	ServerConfPeer := wgtypes.PeerConfig{}
	ServerConfPeer.PublicKey = ClientConf.PrivateKey.PublicKey()

	ClientConfPeer := wgtypes.PeerConfig{}
	ClientConfPeer.PublicKey = ServerConf.PrivateKey.PublicKey()

	serverIp := tunnelIps.IP
	if len(serverIp) == net.IPv4len {
		serverIp[3] = serverIp[3] + 1
	} else {
		serverIp[15] = serverIp[15] + 1
	}
	ServerConf.Address = serverIp.String() + "/" + strings.Split(tunnelIps.String(), "/")[1]

	clientIp := tunnelIps.IP

	copy(clientIp, tunnelIps.IP)

	if len(clientIp) == net.IPv4len {
		clientIp[3] = serverIp[3] + 1
	} else {
		clientIp[15] = clientIp[15] + 1
	}
	ClientConf.Address = clientIp.String() + "/" + strings.Split(tunnelIps.String(), "/")[1]

	wgc.SetServerAddress(ServerConf.Address)
	wgc.SetClientAddress(ClientConf.Address)

	l, err := net.ListenUDP("udp", &net.UDPAddr{Port: 0})
	if err != nil {
		err = fmt.Errorf("can not found an port with err: %s", err)
	}

	port := l.LocalAddr().(*net.UDPAddr).Port
	wgc.SetListenAddr("[::]:" + strconv.Itoa(port))
	l.Close()

	ClientConfPeer.Endpoint = &net.UDPAddr{IP: net.ParseIP(conf.Wireguard.LocalIp), Port: port}
	ServerConf.ListenPort = &port

	ClientConfPeer.AllowedIPs = conf.Wireguard.allowedIps

	netifName := fmt.Sprintf("wg%d", (int(time.Now().Unix())+time.Now().Nanosecond())%1e10)
	wgc.SetNetifName(netifName)
	confName := netifName + ".conf"

	ServerConf.Peers = append(ServerConf.Peers, ServerConfPeer)
	ClientConf.Peers = append(ClientConf.Peers, ClientConfPeer)

	err = ioutil.WriteFile(confName, ServerConf.Parse(), 0644)
	if err != nil {
		err = fmt.Errorf("write %s failed with err: %s", confName, err)
	}

	exec.Command("sudo", "wg-quick", "up", confName).Run()

	_, err = wgc.Save(context.Background())
	if err != nil {
		err = fmt.Errorf("write into db failed with err: %v", err)
	}

	return
}

//func DestroyWireguard(ServerConf wgtypes.Config) error {
//	ws, err := dbCli.WireguardClient.Query().
//		Where(wireguardclient.ServerPrivateKey(ServerConf.PrivateKey.String())).All(context.Background())
//
//}
