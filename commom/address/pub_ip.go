package address

import (
	"github.com/valyala/fasthttp"
)

func GetMyPubIp() string {
	var b []byte
	_, rsp, err := fasthttp.Get(b, "https://api.ip.sb/ip")
	if err != nil {
		return ""
	}
	return string(rsp)
}

func GetMyPubIpv4() string {
	var b []byte
	_, rsp, err := fasthttp.Get(b, "https://api-ipv4.ip.sb/ip")
	if err != nil {
		return ""
	}
	return string(rsp)
}
