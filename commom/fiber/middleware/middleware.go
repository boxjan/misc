package middleware

import (
	"github.com/boxjan/misc/commom/fiber/middleware/logger"
	"github.com/gofiber/fiber/v2"
)

func DefaultWarp(app *fiber.App) {
	app.Use(logger.Logger())
}
