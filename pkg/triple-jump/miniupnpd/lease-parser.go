package miniupnpd

import (
	"bytes"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

// TCP:28043:192.168.3.72:28043:1689610597:NAT-PMP 28043 tcp

type Lease struct {
	RemoteHost             string
	ExternalPort           uint16
	Protocol               string
	InternalPort           uint16
	InternalClient         string
	Enabled                bool
	PortMappingDescription string
	LeaseEndAt             time.Time
}

func (l *Lease) Key() string {
	return strings.ToLower(l.Protocol) + ":" + strconv.Itoa(int(l.ExternalPort)) + ":" + l.InternalClient + ":" + strconv.Itoa(int(l.InternalPort))
}

func LoadLease(path string) ([]Lease, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseLease(b)
}

func parseLease(b []byte) ([]Lease, error) {
	var l []Lease
	ss := bytes.Split(b, []byte{'\n'})
	for _, s := range ss {
		if len(s) == 0 {
			continue
		}
		oneLease, err := parseLeaseLine(string(s))
		if err != nil {
			return nil, err
		}
		l = append(l, oneLease)
	}
	return l, nil
}

func parseLeaseLine(line string) (Lease, error) {
	var lease Lease
	var i int
	var err error
	s := strings.Split(line, ":")
	if len(s) != 6 {
		return lease, errors.New("invalid lease line: " + line)
	}

	lease.Protocol = s[0]

	if i, err = strconv.Atoi(s[1]); err != nil {
		return lease, err
	}
	lease.ExternalPort = uint16(i)

	lease.InternalClient = s[2]

	if i, err = strconv.Atoi(s[3]); err != nil {
		return lease, err
	}
	lease.InternalPort = uint16(i)

	if i, err = strconv.Atoi(s[4]); err != nil {
		return lease, err
	}
	lease.LeaseEndAt = time.Unix(int64(i), 0)

	lease.PortMappingDescription = s[5]

	lease.Enabled = true
	return lease, nil
}

func InLeases(lease Lease, leases []Lease) bool {
	for _, l := range leases {
		if lease.Key() == l.Key() {
			return true
		}
	}
	return false
}
