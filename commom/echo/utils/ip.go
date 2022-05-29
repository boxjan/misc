package utils

import (
	"github.com/gofiber/fiber/v2/utils"
	"net"
	"net/http"
	"strings"
)

// IPs returns an string slice of IP addresses specified in the X-Forwarded-For request header.
func IPs(req *http.Request) (ips []string) {
	header := req.Header.Get("X-Forwarded-For")
	if len(header) == 0 {
		return
	}
	ips = make([]string, strings.Count(header, ",")+1)
	var commaPos, i int
	for {
		commaPos = strings.IndexByte(header, ',')
		if commaPos != -1 {
			ips[i] = strings.Trim(header[:commaPos], " ")
			header, i = header[commaPos+1:], i+1
		} else {
			ips[i] = utils.Trim(header[:commaPos], ' ')
			return
		}
	}
}

func Port(req *http.Request) int {
	addr, err := net.ResolveTCPAddr("", req.RemoteAddr)
	if err != nil {
		return 0
	}
	return addr.Port
}
