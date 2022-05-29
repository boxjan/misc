package middleware

import (
	"github.com/boxjan/misc/commom/fiber/middleware/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
)

func DefaultWarp(app *fiber.App) {
	app.Use(logger.Logger())
	app.Use(pprof.New())
}
