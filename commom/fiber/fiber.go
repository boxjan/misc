package fiber

import (
	"github.com/boxjan/misc/commom/fiber/middleware"
	"github.com/gofiber/fiber/v2"
	"os"
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
	_ = os.Setenv("NO_COLOR", "1")
	app := fiber.New(defaultFiber)

	middleware.DefaultWarp(app)

	return app
}
