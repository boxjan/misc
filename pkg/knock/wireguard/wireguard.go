package wireguard

import (
	"fmt"
	"github.com/valyala/bytebufferpool"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
)

var (
	bp = &bytebufferpool.Pool{}
)

type WgConf struct {
	PrivateKey   wgtypes.Key
	ListenPort   *int
	FirewallMark *int
	Peers        []wgtypes.PeerConfig
}

func (wg WgConf) Parse() []byte {
	w := bp.Get()
	defer bp.Put(w)

	w.WriteString("[Interface]\n")

	fmt.Fprintf(w, "PrivateKey = %s\n", wg.PrivateKey.String())

	if wg.ListenPort != nil {
		fmt.Fprintf(w, "ListenPort = %d\n", *wg.ListenPort)
	}

	if wg.FirewallMark != nil {
		fmt.Fprintf(w, "FwMark = %d\n", *wg.FirewallMark)
	}

	for _, peer := range wg.Peers {
		w.WriteString("[Peer]\n")

		fmt.Fprintf(w, "PublicKey = %s\n", peer.PublicKey.String())

		if peer.Endpoint != nil {
			fmt.Fprintf(w, "Endpoint = %s\n", peer.Endpoint.String())
		}

		if peer.PresharedKey != nil {
			fmt.Fprintf(w, "PresharedKey = %s\n", peer.PresharedKey.String())
		}

		if peer.PersistentKeepaliveInterval != nil {
			fmt.Fprintf(w, "PersistentKeepalive = %d\n", peer.PersistentKeepaliveInterval)
		}

		for _, allowedIp := range peer.AllowedIPs {
			fmt.Fprintf(w, "AllowedIPs = %s\n", allowedIp.String())
		}

		w.WriteByte('\n')
	}

	return w.Bytes()
}

type WgQuickConf struct {
	WgConf

	Address                          *net.IPNet
	DNS                              []net.IP
	MTU                              *int
	Table                            *bool
	PreUp, PostUp, PreDown, PostDown []string
	SaveConfig                       *bool
}

func (wq WgQuickConf) Parse() []byte {
	w := bp.Get()
	defer bp.Put(w)

	// write interface peer
	w.WriteString("[Interface]\n")

	fmt.Fprintf(w, "PrivateKey = %s\n", wq.PrivateKey.String())

	if wq.ListenPort != nil {
		fmt.Fprintf(w, "ListenPort = %d\n", *wq.ListenPort)
	}

	if wq.FirewallMark != nil {
		fmt.Fprintf(w, "FwMark = %d\n", *wq.FirewallMark)
	}

	if wq.Address != nil {
		fmt.Fprintf(w, "Address = %s\n", wq.Address.String())
	}

	for _, dns := range wq.DNS {
		fmt.Fprintf(w, "DNS = %s\n", dns.String())
	}

	if wq.MTU != nil {
		fmt.Fprintf(w, "MTU = %d\n", *wq.MTU)
	}

	if wq.Table != nil && *wq.Table == false {
		fmt.Fprintf(w, "Table = %s\n", "off")
	}

	for _, preUp := range wq.PreUp {
		fmt.Fprintf(w, "PreUp = %s\n", preUp)
	}

	for _, postUp := range wq.PostUp {
		fmt.Fprintf(w, "PostUp = %s\n", postUp)
	}

	for _, preDown := range wq.PreDown {
		fmt.Fprintf(w, "PreDown = %s\n", preDown)
	}

	for _, postDown := range wq.PostDown {
		fmt.Fprintf(w, "PostDown = %s\n", postDown)
	}

	w.WriteByte('\n')

	for _, peer := range wq.Peers {
		w.WriteString("[Peer]\n")

		fmt.Fprintf(w, "PublicKey = %s\n", peer.PublicKey.String())

		if peer.Endpoint != nil {
			fmt.Fprintf(w, "Endpoint = %s\n", peer.Endpoint.String())
		}

		if peer.PresharedKey != nil {
			fmt.Fprintf(w, "PresharedKey = %s\n", peer.PresharedKey.String())
		}

		if peer.PersistentKeepaliveInterval != nil {
			fmt.Fprintf(w, "PersistentKeepalive = %d\n", peer.PersistentKeepaliveInterval)
		}

		for _, allowedIp := range peer.AllowedIPs {
			fmt.Fprintf(w, "AllowedIPs = %s\n", allowedIp.String())
		}

		w.WriteByte('\n')
	}

	return w.Bytes()
}
