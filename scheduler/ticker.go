package scheduler

import (
	"sync"
	"time"
)

type Ticker struct {
	sync.Mutex
	stopped  bool
	interval int64
	offset   int64
	timer    *time.Timer
	C        chan time.Time
	shutdown chan struct{}
}

func NewTicker(interval int64, offset int64) *Ticker {
	t := &Ticker{
		stopped:  true,
		interval: interval,
		offset:   offset,
		C:        make(chan time.Time, 1),
		shutdown: make(chan struct{}),
	}
	return t
}

func (t *Ticker) Ticks() {
	// read from our timer chan until we get a shutdown notfication.
	for {
		select {
		case <-t.shutdown:
			close(t.C)
			return
		case ts := <-t.timer.C:
			t.C <- ts
			t.next()
		}
	}
}

func (t *Ticker) Update(interval int64, offset int64) {
	t.Lock()
	t.interval = interval
	t.offset = offset
	t.Unlock()
}

func (t *Ticker) next() {
	t.Lock()
	if t.stopped {
		// the ticker has been stopped.  No new write will be made to t.C until
		// t.Start() is called again.
		t.Unlock()
		return
	}
	// calculate the number of seconds until our next tick.
	now := time.Now().Unix()
	nextTick := ((t.interval + t.offset) - (now % t.interval)) % t.interval
	if nextTick == 0 {
		nextTick = t.interval
	}

	// set the timer to fire in nextTicks seconds.
	if t.timer == nil {
		t.timer = time.NewTimer(time.Second * time.Duration(nextTick))
		go t.Ticks()
	} else {
		t.timer.Reset(time.Second * time.Duration(nextTick))
	}
	t.Unlock()
}

// start sending ticks on t.C
func (t *Ticker) Start() {
	t.Lock()
	t.stopped = false
	t.Unlock()
	t.next()
}

// stop sending ticks on t.C
func (t *Ticker) Stop() {
	t.Lock()
	t.stopped = true
	t.Unlock()
}

// kill the ticker.  This will close t.C
func (t *Ticker) Delete() {
	t.Stop()
	close(t.shutdown)
}
