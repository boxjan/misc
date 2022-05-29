package middleware

import (
	"github.com/boxjan/misc/commom/echo/middleware/logger"
	"github.com/boxjan/misc/commom/echo/middleware/pprof"
	"github.com/labstack/echo/v4"
)

func DefaultWarp(app *echo.Echo) {
	app.Use(logger.Logger())
	app.Use(pprof.New())
}
