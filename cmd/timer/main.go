package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	. "github.com/boxjan/misc/commom/fiber"
	"github.com/digitorus/timestamp"
	"github.com/gofiber/fiber/v2"
	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
	"net/http"
	"strconv"
	"time"
)

var utcTz = time.UTC

func main() {
	app := DefaultFiber()

	app.Get("/api/v1/time", TimeHandle)
	app.Get("/api/v1/timestamp", TimestampHandle)
	app.Get("/api/v1/tsp", TspHandle)

	klog.Fatal(app.Listen("[::]:8000"))
}

type Ans struct {
	t           time.Time
	TsStr       string `json:"timestamp" yaml:"timestamp" xml:"timestamp"`
	RFC3339Nano string `json:"rfc3339nano" yaml:"rfc3339nano" xml:"rfc3339nano"`
	RFC3339     string `json:"rfc3339" yaml:"rfc3339" xml:"rfc3339"`
	ANSIC       string `json:"ansic" yaml:"ansic" xml:"ansic"`
	Unix        string `json:"unix" yaml:"unix" xml:"unix"`
	Tz          string `json:"tz" yaml:"tz" xml:"tz"`
	Offset      string `json:"offset" yaml:"offset" xml:"offset"`
}

func (a *Ans) Fill() {
	a.TsStr = strconv.FormatInt(a.t.Unix(), 10)
	a.RFC3339Nano = a.t.Format(time.RFC3339Nano)
	a.RFC3339 = a.t.Format(time.RFC3339)
	a.ANSIC = a.t.Format(time.ANSIC)
	a.Unix = a.t.Format(time.UnixDate)
	a.Tz = a.t.Location().String()
	a.Offset = a.t.Format("-07:00")
}

func NowTimeInTz(ctx *fiber.Ctx) Ans {
	tz := ctx.Query("tz", "UTC")
	tzInfo, err := time.LoadLocation(tz)
	if err != nil {
		klog.Infof("ask %s tz, but not found", tz)
		tzInfo = utcTz
		ctx.Status(http.StatusNotFound)
	}

	t := Ans{t: time.Now().In(tzInfo)}
	t.Fill()
	return t
}

func TspHandle(ctx *fiber.Ctx) error {
	return timestamp.ParseError("")
}

func TimeHandle(ctx *fiber.Ctx) error {
	format := ctx.Query("format", "wild")
	now := NowTimeInTz(ctx)

	var data []byte
	var err error

	switch format {
	case "json":
		if data, err = json.Marshal(now); err == nil {
			ctx.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
			return ctx.Send(data)
		}
	case "yaml":
		if data, err = yaml.Marshal(now); err == nil {
			ctx.Set(fiber.HeaderContentType, fiber.MIMETextPlain)
			return ctx.Send(data)
		}
	case "xml":
		ctx.Set(fiber.HeaderContentType, fiber.MIMETextXML)
		if data, err = xml.Marshal(now); err == nil {
			return ctx.Send(data)
		}
	}

	if err != nil {
		klog.Warningf("%s marshal failed with err: %s", format, err)
	}
	ctx.Set(fiber.HeaderContentType, "text/plain")
	return ctx.SendString(now.RFC3339Nano)
}

func TimestampHandle(ctx *fiber.Ctx) error {
	now := time.Now()
	ctx.Set("Content-Type:", "text/plain")
	return ctx.SendString(fmt.Sprintf("%d %d", now.Unix(), now.Nanosecond()))
}
