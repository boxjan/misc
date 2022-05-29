package pprof

import (
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/labstack/echo/v4"
)

func New() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			path := ctx.Path()
			// We are only interested in /debug/pprof routes
			if len(path) < 12 || !strings.HasPrefix(path, "/debug/pprof") {
				return next(ctx)
			}
			switch path {
			case "/debug/pprof/":
				pprof.Index(ctx.Response().Writer, ctx.Request())
			case "/debug/pprof/cmdline":
				pprof.Cmdline(ctx.Response().Writer, ctx.Request())
			case "/debug/pprof/profile":
				pprof.Profile(ctx.Response().Writer, ctx.Request())
			case "/debug/pprof/symbol":
				pprof.Symbol(ctx.Response().Writer, ctx.Request())
			case "/debug/pprof/trace":
				pprof.Trace(ctx.Response().Writer, ctx.Request())
			case "/debug/pprof/allocs":
				pprof.Handler("allocs").ServeHTTP(ctx.Response().Writer, ctx.Request())
			case "/debug/pprof/block":
				pprof.Handler("block").ServeHTTP(ctx.Response().Writer, ctx.Request())
			case "/debug/pprof/goroutine":
				pprof.Handler("goroutine").ServeHTTP(ctx.Response().Writer, ctx.Request())
			case "/debug/pprof/heap":
				pprof.Handler("heap").ServeHTTP(ctx.Response(), ctx.Request())
			case "/debug/pprof/mutex":
				pprof.Handler("mutex").ServeHTTP(ctx.Response().Writer, ctx.Request())
			case "/debug/pprof/threadcreate":
				pprof.Handler("threadcreate").ServeHTTP(ctx.Response().Writer, ctx.Request())
			default:
				// pprof index only works with trailing slash
				if strings.HasSuffix(path, "/") {
					path = strings.TrimRight(path, "/")
				} else {
					path = "/debug/pprof/"
				}

				return ctx.Redirect(http.StatusFound, path)
			}
			return nil
		}
	}
}
