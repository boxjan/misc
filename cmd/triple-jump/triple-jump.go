package main

import (
	"context"
	"github.com/boxjan/misc/pkg/triple-jump/miniupnpd"
	"github.com/boxjan/misc/pkg/triple-jump/upnpc"
	"github.com/huin/goupnp/soap"
	"github.com/kardianos/service"
	"k8s.io/klog/v2"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type tripleJump struct {
	conf Conf

	miniupnpdServiceHandle service.Service

	lastRestartTime time.Time

	miniupnpdConfig miniupnpd.Config

	extInterfaceName string
	extIp            string

	routerExternalIp string

	// merge from forwardExtraList and forwardListInLeaseFileLock
	forwardFullListLock sync.RWMutex
	forwardFullList     []miniupnpd.Lease

	forwardExtraList []miniupnpd.Lease

	forwardListInLeaseFileLock sync.RWMutex
	forwardListInLeaseFile     []miniupnpd.Lease

	forwardListInUpstreamLock sync.RWMutex
	forwardListInUpstream     []miniupnpd.Lease

	miniupnpdLeaseModifyTime time.Time

	routerClient upnpc.RouterClient
	rl           sync.RWMutex
}

func (t *tripleJump) prepareInstance() (err error) {
	t.miniupnpdServiceHandle, err = newServiceHandle(conf.MiniupnpdServiceName)
	if err != nil {
		klog.Errorf("can not new service handle, err: %+v", err)
		return
	}

	t.miniupnpdConfig, err = miniupnpd.LoadConfig(conf.MiniupnpdConfigPath)
	if err != nil {
		klog.Errorf("can not load miniupnpd config, err: %+v", err)
		return
	}

	for _, v := range t.miniupnpdConfig {
		switch v.Key {
		case "ext_ifname":
			t.extInterfaceName = v.Value
		case "ext_ip":
			t.extIp = v.Value
		}
	}

	var extInterface *net.Interface
	extInterface, err = net.InterfaceByName(t.extInterfaceName)
	if err != nil {
		klog.Errorf("can not get interface %s, err: %+v", t.extInterfaceName, err)
		return
	}

	var extInterfaceAddrs []net.Addr
	extInterfaceAddrs, err = extInterface.Addrs()
	if err != nil {
		klog.Errorf("can not get interface %s addrs, err: %+v", t.extInterfaceName, err)
		return
	}

	for _, v := range extInterfaceAddrs {
		if ipNet, ok := v.(*net.IPNet); ok {
			if ipNet.IP.To4() != nil {
				t.routerExternalIp = ipNet.IP.String()
				break
			}
		}
	}
	klog.Infof("will use router external ip: %s", t.routerExternalIp)

	for _, v := range t.conf.ExtraForwards {
		lease := miniupnpd.Lease{
			InternalClient: v.Ip,
			ExternalPort:   v.ExternalPort,
			InternalPort:   v.Port,
			Protocol:       v.Protocol,
			Enabled:        true,
		}

		if lease.ExternalPort == 0 ||
			lease.InternalPort == 0 ||
			(strings.ToLower(lease.Protocol) != "tcp" && strings.ToLower(lease.Protocol) != "udp") {
			klog.Warningf("will not import %+v", v)
			continue
		}

		if lease.InternalClient == "" {
			lease.InternalClient = t.routerExternalIp
		}

		if !miniupnpd.InLeases(lease, t.forwardExtraList) {
			t.forwardExtraList = append(t.forwardExtraList, lease)
		}
	}
	return
}

func (t *tripleJump) checkUpstreamRouter(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case <-ticker.C:
			klog.Info("will get upstream router info")
			t.updateUpstreamRouterExternalIp(ctx)
			t.getUpstreamPortMappingEntries(ctx)
			t.compareAndSyncPortMapping(ctx)
		case <-ctx.Done():
			klog.Info("check upstream router exit")
			return
		}
	}
}

func (t *tripleJump) checkMiniupnpdLeaseFile(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ticker.C:
			klog.Info("will check miniupnpd lease file")
			if t.checkAndReadMiniupnpdLeaseFileIfModTimeChange() { // check if lease file mod changed
				t.mergeFullForwardList()
			}
		case <-ctx.Done():
			klog.Info("check miniupnpd lease file exit")
			return
		}
	}
}

func (t *tripleJump) compareAndSyncPortMapping(ctx context.Context) {

	klog.Infof("will compare and sync port mapping")
	t.forwardListInUpstreamLock.RLock()
	t.forwardListInUpstreamLock.RLock()

	var needAdd, needDel []miniupnpd.Lease
	for _, v := range t.forwardFullList {
		if !miniupnpd.InLeases(v, t.forwardListInUpstream) {
			needAdd = append(needAdd, v)
		}
	}
	for _, v := range t.forwardListInUpstream {
		if !miniupnpd.InLeases(v, t.forwardFullList) {
			needDel = append(needDel, v)
		}
	}

	klog.Infof("full forward list: %+v", t.forwardFullList)
	klog.Infof("forward list in upstream: %+v", t.forwardListInUpstream)
	klog.Infof("need add port mapping: %+v", needAdd)
	klog.Infof("need del port mapping: %+v", needDel)

	t.forwardListInUpstreamLock.RUnlock()
	t.forwardListInUpstreamLock.RUnlock()

	t.rl.RLock()
	defer t.rl.RUnlock()

	var err error
	for _, v := range needAdd {
		for i := 0; i < 3; i++ {
			err = t.routerClient.AddPortMappingCtx(ctx, v.RemoteHost, v.ExternalPort, strings.ToUpper(v.Protocol), v.InternalPort,
				t.routerExternalIp, true, "triple-jump",
				uint32(v.LeaseEndAt.Sub(time.Now()).Seconds()))
			if err != nil {
				klog.Errorf("add port mapping %+v failed, err: %+v", v, err)
			} else {
				klog.Infof("add port mapping %+v success", v)
				break
			}
			time.Sleep(1 * time.Second)
		}
		time.Sleep(1 * time.Second)
	}

	for _, v := range needDel {
		for i := 0; i <= 3; i++ {
			err = t.routerClient.DeletePortMappingCtx(ctx, v.RemoteHost, v.ExternalPort, strings.ToUpper(v.Protocol))
			if err != nil {
				klog.Errorf("delete port mapping %+v failed, err: %+v", v, err)
			} else {
				klog.Infof("delete port mapping %+v success", v)
				break
			}
			time.Sleep(1 * time.Second)
		}
		time.Sleep(1 * time.Second)
	}
}

func (t *tripleJump) updateMiniupnpdConfig() error {
	klog.V(1).Infof("old config: %+v", t.miniupnpdConfig)
	for _, config := range t.miniupnpdConfig {
		if config.Key == "ext_ip" {
			config.Value = t.extIp
		}
	}
	klog.V(1).Infof("new config: %+v", t.miniupnpdConfig)
	return miniupnpd.SaveConfig(conf.MiniupnpdConfigPath, t.miniupnpdConfig)
}

func (t *tripleJump) checkAndReadMiniupnpdLeaseFileIfModTimeChange() (v bool) {
	t.forwardListInLeaseFileLock.Lock()
	defer t.forwardListInLeaseFileLock.Unlock()

	stat, err := os.Stat(conf.MiniupnpdLeasePath)
	if err != nil {
		klog.Errorf("can not stat miniupnpd lease file, err: %+v", err)
		return
	}
	if stat.ModTime().Equal(t.miniupnpdLeaseModifyTime) {
		return
	}

	t.forwardListInLeaseFile, err = miniupnpd.LoadLease(conf.MiniupnpdLeasePath)
	if err != nil {
		klog.Errorf("can not load miniupnpd lease, err: %+v", err)
		return
	}
	t.miniupnpdLeaseModifyTime = stat.ModTime()
	return true
}

func (t *tripleJump) mergeFullForwardList() {
	t.forwardFullListLock.Lock()
	defer t.forwardFullListLock.Unlock()
	t.forwardListInLeaseFileLock.RLock()
	defer t.forwardListInLeaseFileLock.RUnlock()
	var n []miniupnpd.Lease

	for _, v := range t.forwardListInLeaseFile {
		n = append(n, miniupnpd.Lease{
			ExternalPort:   v.ExternalPort,
			Protocol:       v.Protocol,
			InternalPort:   v.ExternalPort,
			InternalClient: t.routerExternalIp,
			Enabled:        true,
			LeaseEndAt:     v.LeaseEndAt,
		})
	}
	n = append(n, t.forwardExtraList...)
	t.forwardFullList = n
}

func (t *tripleJump) getExtIp(ctx context.Context) (string, error) {
	t.rl.RLock()
	defer t.rl.RUnlock()
	return t.routerClient.GetExternalIPAddressCtx(ctx)
}

func (t *tripleJump) updateUpstreamRouterExternalIp(ctx context.Context) {
	var err error
	var ipString string

	for i := 0; i < 3; i++ {
		ipString, err = t.getExtIp(ctx)
		if err != nil {
			klog.Errorf("can not get upstream router external ip, err: %+v", err)
			if i == 0 {
				klog.Infof("will update router client")
				if err = t.updateRouterClient(ctx); err != nil {
					klog.Errorf("can not update router client, err: %+v", err)
					return
				}
			}
		} else {
			break
		}
		time.Sleep(1 * time.Second)
	}

	klog.Infof("upstream router external ip: %s", ipString)

	if net.ParseIP(ipString).To4().String() != strings.TrimSpace(ipString) {
		klog.Errorf("upstream router external ip is not ipv4, ip: %s", ipString)
		return
	}

	if t.extIp != ipString {
		klog.Infof("upstream router external ip changed, old: %s, new: %s", t.extIp, ipString)
		t.extIp = ipString
		klog.Infof("ask for restart miniupnpd")
		go t.updateConfigAndRestartMiniupnpd(ctx)
	}
	return
}

func (t *tripleJump) updateConfigAndRestartMiniupnpd(_ context.Context) {
	if conf.AutoRestartMiniupnpd {
		klog.Infof("will restart miniupnpd")
		if time.Now().Sub(t.lastRestartTime) < 1*time.Second {
			klog.Info("skip restart miniupnpd, last restart time is %v", t.lastRestartTime)
			return
		}
		if err := t.updateMiniupnpdConfig(); err != nil {
			klog.Errorf("update miniupnpd config failed, err: %+v", err)
			return
		}
		t.lastRestartTime = time.Now()
		if err := t.miniupnpdServiceHandle.Restart(); err != nil {
			klog.Errorf("restart miniupnpd failed, err: %+v", err)
			return
		}
	} else {
		klog.Infof("skip restart miniupnpd")
	}
}

func (t *tripleJump) getUpstreamPortMappingEntries(ctx context.Context) {
	t.forwardListInUpstreamLock.Lock()
	defer t.forwardListInUpstreamLock.Unlock()

	klog.Infof("get port mapping entries from upstream router")
	newLeases := []miniupnpd.Lease{}
	var err error
	for i := 0; err == nil; i++ {
		l, e1 := t.getPortMappingEntry(ctx, uint16(i))
		if e1 != nil {
			err = e1
			continue
		}
		newLeases = append(newLeases, l)
	}
	t.forwardListInUpstream = newLeases
	return
}

func (t *tripleJump) getPortMappingEntry(ctx context.Context, index uint16) (miniupnpd.Lease, error) {
	var err error
	var lease miniupnpd.Lease

	for i := 0; i < 3; i++ {
		lease, err = t.getPortMappingEntryOnce(ctx, index)
		if err != nil {
			klog.Errorf("can not get port mapping entry, index: %d, err: %+v", index, err)
			_, ok := err.(*soap.SOAPFaultError)
			if ok {
				return lease, err
			}
		}
		time.Sleep(1 * time.Second)
	}
	return lease, err
}

func (t *tripleJump) getPortMappingEntryOnce(ctx context.Context, index uint16) (miniupnpd.Lease, error) {
	var err error
	var lease miniupnpd.Lease
	remoteHost, extPort, protocol, intPort, intClient, enabled, desc, leaseDuration, err := t.routerClient.GetGenericPortMappingEntryCtx(ctx, index)
	if err != nil {
		return lease, err
	}
	lease = miniupnpd.Lease{
		RemoteHost:             remoteHost,
		ExternalPort:           extPort,
		Protocol:               protocol,
		InternalPort:           intPort,
		InternalClient:         intClient,
		Enabled:                enabled,
		PortMappingDescription: desc,
		LeaseEndAt:             time.Now().Add(time.Duration(leaseDuration) * time.Second),
	}
	return lease, nil
}

func (t *tripleJump) updateRouterClient(ctx context.Context) (err error) {
	t.rl.Lock()
	defer t.rl.Unlock()

	var r upnpc.RouterClient
	for i := 5; i > 0; i-- {
		if r, err = upnpc.PickRouterClient(ctx); err != nil {
			klog.Infof("pick router client failed err: %+v", err)
		} else {
			t.routerClient = r
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return
}
