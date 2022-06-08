package wireguard

import (
	"fmt"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
	"time"
)

var (
	False = false
)

// NewWgQuickConfPair will generate server and client wireguard config
func NewWgQuickConfPair(serverIp, clientIp *net.IPNet, listen *net.UDPAddr, keepaliveInterval int,
	clientAllowedIps ...net.IPNet) (server, client *WgQuickConf, err error) {

	server, client = &WgQuickConf{}, &WgQuickConf{}

	server.Table = &False
	server.ListenPort = &listen.Port
	server.Address = serverIp

	client.Address = clientIp

	server.PrivateKey, err = wgtypes.GeneratePrivateKey()
	if err != nil {
		err = fmt.Errorf("generate private key failed with err %v", err)
		return
	}
	client.PrivateKey, err = wgtypes.GeneratePrivateKey()
	if err != nil {
		err = fmt.Errorf("generate private key failed with err %v", err)
		return
	}

	persistentKeepaliveInterval := 10 * time.Second
	if keepaliveInterval != 0 {
		persistentKeepaliveInterval = time.Duration(keepaliveInterval) * time.Second
	}

	server.Peers = append(server.Peers, wgtypes.PeerConfig{
		PublicKey:                   client.PrivateKey.PublicKey(),
		PersistentKeepaliveInterval: &persistentKeepaliveInterval,
		AllowedIPs:                  nil,
	})

	client.Peers = append(client.Peers, wgtypes.PeerConfig{
		PublicKey:                   server.PrivateKey.PublicKey(),
		PersistentKeepaliveInterval: &persistentKeepaliveInterval,
		Endpoint:                    listen,
		AllowedIPs:                  clientAllowedIps,
	})

	return
}
