package game

import (
	"sync"
	"time"
)

// TimedEventType identifies the category of timed event.
type TimedEventType string

const (
	EventBuildComplete TimedEventType = "build"
	EventTrainComplete TimedEventType = "train"
	EventGatherExpire  TimedEventType = "gather"
)

// TimedEvent represents a single scheduled event.
type TimedEvent struct {
	ID       string
	Type     TimedEventType
	RID      int64
	FireAt   time.Time
	Callback func()
}

// TimingWheel manages all timed events with a single ticker.
type TimingWheel struct {
	mu      sync.Mutex
	events  map[string]*TimedEvent
	ticker  *time.Ticker
	done    chan struct{}
	running bool
}

// NewTimingWheel creates a new TimingWheel.
func NewTimingWheel() *TimingWheel {
	return &TimingWheel{
		events: make(map[string]*TimedEvent),
	}
}

// Start begins the timing wheel's check loop.
func (tw *TimingWheel) Start() {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.running {
		return
	}

	tw.running = true
	tw.ticker = time.NewTicker(1 * time.Second)
	tw.done = make(chan struct{})
	go tw.loop()
}

// Stop halts the timing wheel.
func (tw *TimingWheel) Stop() {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if !tw.running {
		return
	}

	tw.running = false
	if tw.ticker != nil {
		tw.ticker.Stop()
	}
	close(tw.done)
}

// RegisterEvent adds or replaces a timed event.
func (tw *TimingWheel) RegisterEvent(event *TimedEvent) {
	if event == nil {
		return
	}

	tw.mu.Lock()
	defer tw.mu.Unlock()
	tw.events[event.ID] = event
}

// CancelEvent removes a timed event by ID.
func (tw *TimingWheel) CancelEvent(eventID string) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	delete(tw.events, eventID)
}

func (tw *TimingWheel) loop() {
	for {
		select {
		case <-tw.done:
			return
		case <-tw.ticker.C:
			tw.fireDue()
		}
	}
}

func (tw *TimingWheel) fireDue() {
	tw.mu.Lock()
	now := time.Now()
	var due []*TimedEvent
	for id, evt := range tw.events {
		if !now.Before(evt.FireAt) {
			due = append(due, evt)
			delete(tw.events, id)
		}
	}
	tw.mu.Unlock()

	for _, evt := range due {
		if evt.Callback != nil {
			evt.Callback()
		}
	}
}
