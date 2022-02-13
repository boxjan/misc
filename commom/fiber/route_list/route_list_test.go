package route_list

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"testing"
)

func TestRouteList(t *testing.T) {
	app := fiber.New()

	app.Get("/", defaultGet)

	app.Get("/g1", defaultGet)
	app.Get("/g2", defaultGet)

	app.Post("/post", defaultPost)
	app.Post("/post/:index", defaultPost)

	g := app.Group("/group")
	g.Get("/get", defaultGet)
	g.Get("/get/:index", defaultGet)
	g.Post("post", defaultPost)
	app.Group("/group").Post("post2", defaultPost)

	g = app.Group("/gr2", defaultMiddleware)
	g.Get("/get", defaultGet)
	g.Get("/get/:index", defaultGet)
	g.Post("post", defaultPost)

	app.Use("/m1m1", defaultMiddleware).Get("get", defaultGet)
	app.Use("/m1m2", defaultMiddleware).Get("get", defaultGet)
	app.Use("/m2m2", defaultMiddleware).Get("get", defaultGet)

	app.Get("/g2", defaultGet2)
	c, routeList := RouteList(app)
	fmt.Println(c, routeList)
}

func defaultMiddleware(ctx *fiber.Ctx) error {
	return nil
}

func defaultGet(ctx *fiber.Ctx) error {
	return nil
}

func defaultGet2(ctx *fiber.Ctx) error {
	return nil
}

func defaultPost(ctx *fiber.Ctx) error {
	return nil
}
