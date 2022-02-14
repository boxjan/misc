package main

import (
	. "github.com/boxjan/misc/commom/fiber"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
	"strconv"
	"sync"
	"time"
)

type Gunshots struct {
	sync.Mutex

	name string

	expectNum  int32
	now        int32
	expireTime time.Time
	bang       []chan struct{}
}

func (g *Gunshots) Bang() {
	klog.Info("will bang !")
	g.Lock()
	defer g.Unlock()
	for k := range g.bang {
		g.bang[k] <- struct{}{}
	}
	allShots.LoadAndDelete(g.name)
}

func (g *Gunshots) Expire() {
	t := time.NewTicker(g.expireTime.Sub(time.Now()))
	select {
	case <-t.C:
	case <-g.bang[0]:
	}
	g.Bang()
}

var allShots sync.Map

func main() {
	app := DefaultFiber()

	app.Get("/api/v1/gun/:game", gunHandle)

	klog.Fatal(app.Listen("[::]:8001"))
}

func gunHandle(ctx *fiber.Ctx) error {
	game := ctx.Params("game", "")
	expectNum, _ := strconv.Atoi(ctx.Query("num", "0"))

	if game == "" {
		ctx.Status(404)
		return ctx.SendString("need game id")
	}

	if expectNum <= 0 {
		ctx.Status(404)
		return ctx.SendString("num id is too small")
	}

	shotPtr, loaded := allShots.LoadOrStore(game, NewShot(expectNum))
	shot := shotPtr.(*Gunshots)

	bang := make(chan struct{}, 1)

	shot.Lock()
	shot.bang = append(shot.bang, bang)
	shot.now += 1
	if shot.expectNum == shot.now {
		go shot.Bang()
	}
	shot.Unlock()

	_ = <-bang

	if !loaded {
		allShots.Delete(game)
	}

	return ctx.SendString("go!")
}

func NewShot(expectNum int) *Gunshots {
	g := &Gunshots{
		Mutex:      sync.Mutex{},
		expectNum:  int32(expectNum),
		expireTime: time.Now().Add(1 * time.Hour),
		bang:       make([]chan struct{}, 0),
	}
	g.bang = append(g.bang, make(chan struct{}, 1))
	go g.Expire()
	return g
}
