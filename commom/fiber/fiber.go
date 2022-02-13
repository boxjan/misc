package fiber

import (
	"github.com/boxjan/misc/commom/fiber/middleware"
	"github.com/gofiber/fiber/v2"
)

var defaultFiber = fiber.Config{

	Prefork:                 false,
	ServerHeader:            "fiber",
	Network:                 fiber.NetworkTCP,
	EnableTrustedProxyCheck: true,
	TrustedProxies:          []string{"127.0.0.0/8"},
	EnablePrintRoutes:       true,
}

func DefaultFiber() *fiber.App {
	app := fiber.New()

	middleware.DefaultWarp(app)

	return app
}
