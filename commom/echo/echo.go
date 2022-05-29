package echo

import (
	"github.com/boxjan/misc/commom/echo/middleware"
	"github.com/labstack/echo/v4"
	"net/http"
)

func DefaultEcho() *echo.Echo {
	app := echo.New()
	app.IPExtractor = echo.ExtractIPFromXFFHeader(
		echo.TrustLoopback(true),
	)

	middleware.DefaultWarp(app)

	app.GET("/healthz", func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "OK")
	}).Name = "healthy-check"
	return app
}
