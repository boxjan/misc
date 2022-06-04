package main

import (
	"github.com/gofiber/fiber/v2"
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

}
