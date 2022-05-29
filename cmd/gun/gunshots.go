package main

import (
	"k8s.io/klog/v2"
	"sync"
	"time"
)

type Gunshots struct {
	sync.RWMutex

	name string

	expectNum int
	now       int32
	expireAt  time.Time
	fire      chan struct{}
	started   bool
	startedAt time.Time
	user      map[string]chan int
}

func (g *Gunshots) Bang() {
	<-g.fire
	g.Lock()
	defer g.Unlock()
	for k := range g.user {
		g.user[k] <- 0
	}
}

func (g *Gunshots) Cancel() {
	g.Lock()
	defer g.Unlock()

	for k := range g.user {
		g.user[k] <- -2
	}
}

func (g *Gunshots) CheckRegisteredCount() {
	for range time.NewTicker(time.Second).C {
		if g.started {
			return
		}
		g.Lock()
		if len(g.user) == g.expectNum+1 {
			g.started = true
			g.startedAt = time.Now()
			g.fire <- struct{}{}
		}
		g.Unlock()
	}
}

func NewShot(game string, expectNum int) *Gunshots {
	g := &Gunshots{
		name:      game,
		RWMutex:   sync.RWMutex{},
		expectNum: expectNum,
		expireAt:  time.Now().Add(1 * time.Hour),
		user:      map[string]chan int{},
		fire:      make(chan struct{}, 1),
	}
	go g.CheckRegisteredCount()
	go g.Bang()
	return g
}

func checkAllShots() {
	for range time.NewTicker(1 * time.Minute).C {
		klog.V(1).Info("check all shots status")
		allGames.Range(func(key, value interface{}) bool {
			klog.V(1).Infof("check")
			shot, vok := value.(*Gunshots)
			if !vok {
				klog.Warningf("ahh, what's %+v %+v this, it must not a shot", key, value)
				allGames.Delete(key)
			}

			if shot.started {
				if time.Now().Sub(shot.startedAt) > 1*time.Hour {
					klog.Infof("game %s %+v have started 1 hour ago, forget it now", key, value)
					allGames.Delete(key)
				}
			} else if time.Now().Sub(shot.expireAt) > 0 {
				klog.Infof("game %s %+v not start too long time, forget it now", key, value)
				shot.Cancel()
				allGames.Delete(key)
			}
			return true
		})
	}
}

func cancelAllShots() {
	allGames.Range(func(key, value interface{}) bool {
		klog.Warning("will cancel %s", key)
		shot, vok := value.(*Gunshots)
		if vok {
			shot.Cancel()
		}
		return true
	})
}
