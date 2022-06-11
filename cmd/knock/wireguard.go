package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/boxjan/misc/pkg/knock/ent"
	"github.com/boxjan/misc/pkg/knock/ent/wireguardclient"
	"github.com/boxjan/misc/pkg/knock/wireguard"
	"golang.zx2c4.com/wireguard/wgctrl"
	"k8s.io/klog/v2"
	"net"
	"time"
)

func wireguardBackground() {
	var wgCli *wgctrl.Client
	var err error

	for range time.NewTicker(5 * time.Second).C {
		if wgCli == nil {
			if wgCli, err = wgctrl.New(); err != nil {
				klog.Warningf("new wireguard client failed with err: %v", err)
			}
		}

		devices, err := wgCli.Devices()
		if err != nil {
			klog.Warningf("get all device failed with err: %v", err)
			continue
		}

		for _, device := range devices {
			wgc, dbErr := dbCli.WireguardClient.Query().Where(
				wireguardclient.NetifName(device.Name)).First(context.Background())
			if dbErr != nil {
				klog.Warningf("get client info %s failed with err: %v", device.Name, dbErr)
				continue
			}
			if wgc == nil {
				klog.Warningf("no found device: %s maybe not alloc by knock", device.Name)
				continue
			}
			for _, peer := range device.Peers {
				if peer.Endpoint == nil {
					if time.Now().Sub(wgc.CreatedAt) >= 3*time.Minute {
						klog.Infof("will shutdown %s, reason: long time no connect")
						if err := destroyWireguard(wgc); err != nil {
							klog.Warningf("shutdown %s failed with err: %s",
								device.Name)
						}
					}
					continue
				} else {
					if peer.Endpoint.String() != wgc.PeerAddr {
						klog.Warningf("device: %s peer from %s -> %s", device.Name, wgc.PeerAddr, peer.Endpoint.String())
						if _, err := wgc.Update().
							SetPeerAddr(peer.Endpoint.String()).
							Save(context.Background()); err != nil {
							klog.Warningf("save %s failed with err: %s", device.Name, err)
						}
					}
				}

				if wgc.TransmitBytes == uint64(peer.TransmitBytes) || wgc.ReceiveBytes == uint64(peer.ReceiveBytes) {
					if time.Now().Sub(wgc.UpdatedAt) >= time.Minute {
						klog.Infof("will shutdown %s, reason: long time no transit")
						if _, err := wgc.Update().
							SetTransmitBytes(uint64(peer.TransmitBytes)).
							SetReceiveBytes(uint64(peer.ReceiveBytes)).
							Save(context.Background()); err != nil {
							klog.Warningf("save %s failed with err: %s", device.Name, err)
						}
						if err := destroyWireguard(wgc); err != nil {
							klog.Warningf("shutdown %s failed with err: %s",
								device.Name)
						}
					}

				} else {
					if _, err := wgc.Update().
						SetTransmitBytes(uint64(peer.TransmitBytes)).
						SetReceiveBytes(uint64(peer.ReceiveBytes)).
						Save(context.Background()); err != nil {
						klog.Warningf("save %s failed with err: %s", device.Name, err)
					}
				}

				klog.V(1).Infof("device %s rx: %s, tx: %s", peer.ReceiveBytes, peer.TransmitBytes)
			}
			klog.V(2).Infof("device %s, info: %+v", device)
		}
	}
}

func NewWireguard(identify string) (ClientConf *wireguard.WgQuickConf, err error) {

	var wgc = dbCli.WireguardClient.Create().SetIdentify(identify)
	var ServerConf *wireguard.WgQuickConf

	var tunnelIps *net.IPNet
	tunnelIps, err = cidrSet.AllocateNext()
	if err != nil {
		klog.Warning("alloc ip block failed with err: %v", err)
		return
	}

	serverIps := *tunnelIps
	serverIps.IP = make([]byte, len(tunnelIps.IP))
	copy(serverIps.IP, tunnelIps.IP)
	if len(serverIps.IP) == net.IPv4len {
		serverIps.IP[3] = serverIps.IP[3] + 1
	} else {
		serverIps.IP[15] = serverIps.IP[15] + 1
	}

	clientIps := *tunnelIps
	clientIps.IP = make([]byte, len(tunnelIps.IP))
	copy(clientIps.IP, tunnelIps.IP)
	if len(clientIps.IP) == net.IPv4len {
		clientIps.IP[3] = clientIps.IP[3] + 2
	} else {
		clientIps.IP[15] = clientIps.IP[15] + 2
	}

	l, err := net.ListenUDP("udp", &net.UDPAddr{Port: 0})
	if err != nil {
		err = fmt.Errorf("can not found an port with err: %s", err)
	}
	addr := l.LocalAddr().(*net.UDPAddr)
	addr.IP = net.ParseIP(conf.Wireguard.LocalIp)
	l.Close()

	ServerConf, ClientConf, err = wireguard.NewWgQuickConfPair(&serverIps, &clientIps, addr, 10,
		conf.Wireguard.allowedIps...)

	wgc.SetAllocCidr(tunnelIps.String())
	wgc.SetListenAddr(addr.String())
	wgc.SetServerAddress(serverIps.IP.String())
	wgc.SetClientAddress(clientIps.IP.String())
	wgc.SetServerPrivateKey(hex.EncodeToString(ServerConf.PrivateKey[:]))
	wgc.SetClientPrivateKey(hex.EncodeToString(ClientConf.PrivateKey[:]))

	netifName := fmt.Sprintf("wg%d", (int(time.Now().Unix())+time.Now().Nanosecond())%1e10)
	wgc.SetNetifName(netifName)

	err = wireguard.SetUpWireguardLink(netifName, ServerConf)

	_, err = wgc.Save(context.Background())
	if err != nil {
		err = fmt.Errorf("write into db failed with err: %v", err)
	}

	return
}

func destroyWireguard(wgc *ent.WireguardClient) error {
	err := wireguard.ShutdownWgQuickLink(wgc.NetifName)
	if err != nil {
		klog.Error("shutdown device %s failed with err", wgc.NetifName)
	}
	wgc.Expired = true
	wgc.DestroyedAt = time.Now()
	if _, err := wgc.Update().
		SetExpired(true).
		SetDestroyedAt(time.Now()).
		Save(context.Background()); err != nil {
		klog.Warningf("save %s failed with err: %s", wgc.NetifName, err)
	}
	if _, cidr, e := net.ParseCIDR(wgc.AllocCidr); e != nil {
		klog.Warningf("release alloc %s with err: %s", wgc.AllocCidr, cidrSet.Release(cidr))
	} else {
		klog.V(1).Infof("release alloc %s", wgc.AllocCidr)
	}
	return nil
}
