package main

import (
	"github.com/boxjan/misc/commom/cmd"
	. "github.com/boxjan/misc/commom/echo"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
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
	app := DefaultEcho()

	apiV1 := app.Group("/api/v1/")

	apiV1.POST("gun/:game/:partner", newGameHandle).Name = "api-v1/new-game"
	apiV1.GET("gun/:game/:partner", gameHandle).Name = "api-v1/join-game"

	klog.Fatal(app.Start(conf.Addr))
}

func newGameHandle(ctx echo.Context) error {
	g := ctx.Param("game")
	if len(g) == 0 {
		return ctx.String(http.StatusTeapot, "plz tell me the new game name")
	}

	sth, ok := allGames.Load(g)
	if !ok {
		num, e := strconv.Atoi(ctx.Param("partner"))
		if e != nil || num <= 0 {
			klog.Warningf("someone ask a new game with %+v participant", ctx.QueryParam("num"))
			return ctx.String(http.StatusTeapot, "plz tell me the new game expect number of participants")
		}

		shot := NewShot(g, num)
		allGames.Store(g, shot)
		return ctx.String(http.StatusOK, "the new game is ready for join")
	}

	game, vok := sth.(*Gunshots)
	if !vok {
		klog.Errorf("ahh, what's %+v this, it must not a shot", sth)
		return ctx.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if game.started {
		return ctx.String(http.StatusGone, "the game is start, you are too late")
	}

	return ctx.String(http.StatusOK, "the game is waiting for you")
}

func gameHandle(ctx echo.Context) error {
	game := ctx.Param("game")
	partner := ctx.Param("partner")

	if len(game) == 0 {
		return ctx.String(http.StatusTeapot, "plz tell me the game name")
	}

	sth, ok := allGames.Load(game)
	if !ok {
		return ctx.String(http.StatusTeapot, "can not find the game, check api list then register it")
	}

	shot, vok := sth.(*Gunshots)
	if !vok {
		klog.Errorf("ahh, what's %+v this, it must not a shot", sth)
		return ctx.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if shot.started {
		return ctx.String(http.StatusGone, "the game is start, you are too late")
	}

	hold := make(chan int, 1)
	shot.Lock()
	if ch, ok := shot.user[partner]; ok {
		klog.Warningf("game: %s partner: %s have join the game", game, partner)
		ch <- -1
	}
	shot.user[partner] = hold
	shot.Unlock()

	rsp := ctx.Response()
	keepAliveTicker := time.NewTicker(5 * time.Second).C
	for {
		select {
		case <-keepAliveTicker:
			_, _ = rsp.Write([]byte("still wait at: " + time.Now().Format(time.RFC3339) + "\n"))
		case t := <-hold:
			switch t {
			case -1:
				_, _ = rsp.Write([]byte("some one join the game with same name, bye"))
			case -2:
				_, _ = rsp.Write([]byte("cancel this game"))
			default:
				_, _ = rsp.Write([]byte("It's time to go"))
			}
			return nil
		}
		rsp.Flush()
	}
}
