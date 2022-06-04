package main

import (
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
	"net/http"
)

func NewWireguardHandle(ctx *fiber.Ctx) error {
	var identify string
	if conf.Http.IdentityHeader != "" {
		identify = ctx.Get(conf.Http.IdentityHeader)
		if identify == "" {
			ctx.Status(http.StatusUnauthorized)
			return ctx.SendString("could not get identity")
		}
	}
	cli, err := NewWireguard(identify)
	if err != nil {
		klog.Error(err)
		ctx.Status(500)
		return ctx.SendString("failed create wg tunnel")
	}

	return ctx.Send(cli.Parse())
}
