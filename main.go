package main

import (
	"fmt"
	"sync"
)

type Subject[T any] interface {
	Subscribe(o Observer[T])
	Unsubscribe(o Observer[T])
	Notify(e *T)
}

type Observer[T any] interface {
	Update(e *T)
}

type LocationObserver struct {
	ID        int
	latitude  float64
	longitude float64
	event     chan *TelemetryEvent
	wg        *sync.WaitGroup
}

func (o *LocationObserver) Update(e *TelemetryEvent) {
	o.event <- e
}

func (o *LocationObserver) handleUpdate() {
	e := <-o.event
	o.latitude = e.Latitude
	o.longitude = e.Longitude

	fmt.Printf("Location Observer %d: Latitude: %f, Longitude: %f\n", o.ID, o.latitude, o.longitude)
}

func (o *LocationObserver) Run() {
	o.wg.Add(1)

	go func() {
		defer o.wg.Done()
		o.handleUpdate()
	}()
}

type TemperatureObserver struct {
	ID          int
	temperature float64
	event       chan *TelemetryEvent
	wg          *sync.WaitGroup
}

func (o *TemperatureObserver) Update(e *TelemetryEvent) {
	o.event <- e
}

func (o *TemperatureObserver) handleUpdate() {
	e := <-o.event
	o.temperature = e.Temperature

	fmt.Printf("Location Observer %d: Temperature: %f\n", o.ID, o.temperature)
}

func (o *TemperatureObserver) Run() {
	o.wg.Add(1)

	go func() {
		defer o.wg.Done()
		o.handleUpdate()
	}()
}

func newObserver[T TemperatureObserver | LocationObserver](wg *sync.WaitGroup, id int) Observer[TelemetryEvent] {
	var x T
	switch (interface{}(x)).(type) {
	case TemperatureObserver:
		o := TemperatureObserver{
			ID:    id,
			wg:    wg,
			event: make(chan *TelemetryEvent),
		}

		o.Run()

		return &o
	case LocationObserver:
		o := LocationObserver{
			ID:    id,
			wg:    wg,
			event: make(chan *TelemetryEvent),
		}

		o.Run()

		return &o
	}

	return nil
}

type TelemetryEvent struct {
	Latitude    float64
	Longitude   float64
	Temperature float64
}

type TelemetrySubject struct {
	observers []Observer[TelemetryEvent]
}

func (s *TelemetrySubject) Subscribe(o Observer[TelemetryEvent]) {
	s.observers = append(s.observers, o)
}

func (s *TelemetrySubject) Notify(e *TelemetryEvent) {
	for _, o := range s.observers {
		o.Update(e)
	}
}

func main() {
	wg := sync.WaitGroup{}

	lo1 := newObserver[LocationObserver](&wg, 1)
	to1 := newObserver[TemperatureObserver](&wg, 2)
	lo2 := newObserver[LocationObserver](&wg, 3)
	to2 := newObserver[TemperatureObserver](&wg, 4)

	ts := TelemetrySubject{
		observers: make([]Observer[TelemetryEvent], 0),
	}

	ts.Subscribe(lo1)
	ts.Subscribe(to1)
	ts.Subscribe(lo2)
	ts.Subscribe(to2)

	ts.Notify(&TelemetryEvent{
		Temperature: 100,
		Latitude:    50,
		Longitude:   50,
	})

	wg.Wait()
}
