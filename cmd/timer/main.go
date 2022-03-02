package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/boxjan/misc/commom/cmd"
	. "github.com/boxjan/misc/commom/fiber"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"k8s.io/klog/v2"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"
)

var utcTz = time.UTC

var (
	rootCmd = cmd.QuickCobraRun("timer", run)

	configPath = &cmd.ConfigPath

	conf = Config{}
)

func init() {
	initCmd()
}

func initCmd() {
}

func main() {
	if e := recover(); e != nil {
		trace := debug.Stack()
		klog.Errorf("catch panic: %v\nstack: %s", e, trace)
		return
	}

	rootCmd.PreRunE = preRunE

	if err := rootCmd.Execute(); err != nil {
		klog.Error(err)
	}
}

func preRunE(cmd *cobra.Command, args []string) error {
	if d, err := ioutil.ReadFile(*configPath); err != nil {
		klog.Warningf("can not load conf, will use default config")
		conf = defaultConfig
		return nil
	} else if err := yaml.Unmarshal(d, &conf); err != nil {
		klog.Warningf("can not load conf, will use default config")
		conf = defaultConfig
		return nil
	}
	return nil
}

func run(cmd *cobra.Command, args []string) {
	app := DefaultFiber()

	apiV1 := app.Group("/api/v1").Name("api-v1/")

	/*
		/api/v1/time
		/api/v1/timestamp
	*/
	apiV1.Get("time", TimeHandle).Name("time")
	apiV1.Get("timestamp", TimestampHandle).Name("timestamp")

	klog.Fatal(app.Listen(conf.Addr))
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
	return t
}

func TimeHandle(ctx *fiber.Ctx) error {
	format := ctx.Query("format", "wild")
	now := NowTimeInTz(ctx)

	var data []byte
	var err error

	switch format {
	case "json":
		now.Fill()
		if data, err = json.Marshal(now); err == nil {
			ctx.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
			return ctx.Send(data)
		}
	case "xml":
		now.Fill()
		ctx.Set(fiber.HeaderContentType, fiber.MIMETextXML)
		if data, err = xml.Marshal(now); err == nil {
			return ctx.Send(data)
		}
	default:
		ctx.Set(fiber.HeaderContentType, fiber.MIMETextPlain)
		for k, f := range conf.ExtraFormat {
			if strings.ToUpper(format) == strings.ToUpper(k) {
				return ctx.SendString(now.t.Format(f))
			}
		}

		for k, f := range innerFormat {
			if strings.ToUpper(format) == strings.ToUpper(k) {
				return ctx.SendString(now.t.Format(f))
			}
		}

	}

	if err != nil {
		klog.Warningf("%s marshal failed with err: %s", format, err)
	}
	return ctx.SendString(now.t.Format(time.RFC3339Nano))
}

func TimestampHandle(ctx *fiber.Ctx) error {
	now := time.Now()
	ctx.Set("Content-Type:", "text/plain")
	return ctx.SendString(fmt.Sprintf("%d %d", now.Unix(), now.Nanosecond()))
}
