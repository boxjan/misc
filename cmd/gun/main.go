package main

import (
	"github.com/boxjan/misc/commom/cmd"
	. "github.com/boxjan/misc/commom/fiber"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"k8s.io/klog/v2"
	"net/http"
	"runtime/debug"
	"strconv"
	"sync"
	"time"
)

// After every one ready, bang and run

var (
	rootCmd    = cmd.QuickCobraRun("timer", run)
	configPath = &cmd.ConfigPath
	conf       = Config{}

	allGames sync.Map
)

func init() {
	go checkAllShots()
}

func main() {
	defer func() {
		if e := recover(); e != nil {
			trace := debug.Stack()
			klog.Errorf("catch panic: %v\nstack: %s", e, trace)
			return
		}
	}()
	defer cancelAllShots()

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

	apiV1.Post("gun/:game", newGameHandle)
	apiV1.Get("gun/:game/:partner", gameHandle)

	klog.Fatal(app.Listen(conf.Addr))
}

func newGameHandle(ctx *fiber.Ctx) error {
	g := ctx.Params("game", "")
	if len(g) == 0 {
		ctx.Status(http.StatusTeapot)
		return ctx.SendString("plz tell me the new game name")
	}

	sth, ok := allGames.Load(g)
	if !ok {
		num, e := strconv.Atoi(ctx.Query("num", "0"))
		if e != nil || num <= 0 {
			klog.Warningf("someone ask a new game with %+v participant", ctx.Query("num", "0"))
			ctx.Status(http.StatusTeapot)
			return ctx.SendString("plz tell me the new game expect number of participants")
		}

		shot := NewShot(g, num)
		allGames.Store(g, shot)
		ctx.Status(http.StatusOK)
		return ctx.SendString("the new game is ready for join")
	}

	game, vok := sth.(*Gunshots)
	if !vok {
		klog.Errorf("ahh, what's %+v this, it must not a shot", sth)
		ctx.Status(http.StatusInternalServerError)
		return ctx.SendString(http.StatusText(http.StatusInternalServerError))
	}

	if game.started {
		ctx.Status(http.StatusGone)
		return ctx.SendString("the game is start, you are too late")
	}

	ctx.Status(http.StatusOK)
	return ctx.SendString("the game is waiting for you")

}

func gameHandle(ctx *fiber.Ctx) error {
	game := ctx.Params("game", "")
	partner := ctx.Params("partner", "")

	if len(game) == 0 {
		ctx.Status(http.StatusTeapot)
		return ctx.SendString("plz tell me the game name")
	}

	sth, ok := allGames.Load(game)
	if !ok {
		ctx.Status(http.StatusTeapot)
		return ctx.SendString("can not find the game, check api list then register it")
	}

	shot, vok := sth.(*Gunshots)
	if !vok {
		klog.Errorf("ahh, what's %+v this, it must not a shot", sth)
		ctx.Status(http.StatusInternalServerError)
		return ctx.SendString(http.StatusText(http.StatusInternalServerError))
	}

	hold := make(chan int, 1)
	shot.Lock()
	if ch, ok := shot.user[partner]; ok {
		klog.Warningf("game: %s partner: %s have join the game", game, partner)
		ch <- -1
	}
	shot.user[partner] = hold
	shot.Unlock()

	pr, pw := io.Pipe()
	go func() {
		keepAliveTicker := time.NewTicker(10 * time.Second).C
		for {
			select {
			case <-keepAliveTicker:
				_, _ = pw.Write([]byte("still wait at: " + time.Now().Format(time.RFC3339) + "\n"))
			case t := <-hold:
				switch t {
				case -1:
					_, _ = pw.Write([]byte("some one join the game with same name, bye"))
				case -2:
					_, _ = pw.Write([]byte("cancel this game"))
				default:
					_, _ = pw.Write([]byte("It's time to go"))
				}
				_ = pw.Close()
				return
			}
		}
	}()

	return ctx.SendStream(pr)
}
