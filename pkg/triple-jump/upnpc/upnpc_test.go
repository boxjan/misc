package upnpc

import (
	"context"
	"github.com/boxjan/misc/pkg/triple-jump/miniupnpd"
	"testing"
	"time"
)

func TestGetPortMappingEntries(t *testing.T) {
	cli, err := PickRouterClient(context.Background())
	if err != nil {
		t.Fatalf("pick up client err: %s", err)
	}

	ip, err := cli.GetExternalIPAddressCtx(context.Background())
	if err != nil {
		t.Fatalf("get ext ip err: %s", err)
	}

	t.Logf("ext ip: %s", ip)

	for i := 0; i < 10; i++ {
		remoteHost, extPort, protocol, intPort, intClient, enabled, desc, leaseDuration, err :=
			cli.GetGenericPortMappingEntryCtx(context.Background(), uint16(i))
		if err != nil {
			t.Errorf("can not get port mapping entry, index: %d, err: %+v", i, err)
			continue
		}

		lease := miniupnpd.Lease{
			RemoteHost:             remoteHost,
			ExternalPort:           extPort,
			Protocol:               protocol,
			InternalPort:           intPort,
			InternalClient:         intClient,
			Enabled:                enabled,
			PortMappingDescription: desc,
			LeaseEndAt:             time.Now().Add(time.Duration(leaseDuration) * time.Second),
		}
		t.Logf("%+v", lease)
	}
}
