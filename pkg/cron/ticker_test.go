package cron

import (
	"context"
	"testing"
	"time"
)

func TestNewTicker(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(2)*time.Second)
	defer cancel()

	chTrigger := make(chan time.Time)
	ticker := NewTicker(ctx, EverySecond(), chTrigger)

	var err error
	if err = ticker.stop(); err == nil { // Ticker should not be started yet
		t.Errorf("ticker stop should return an error")
	}

	if err = ticker.start(); err != nil {
		t.Errorf("ticker start should not return an error, received %v", err)
	}

	if err = ticker.start(); err == nil { // try starting the goroutine again
		t.Errorf("ticker start should return an error")
	}

	var exit bool
	for {
		if exit {
			break
		}

		select {
		case <-ctx.Done():
			t.Errorf("ticker should have been triggered")
			exit = true
		case <-chTrigger:
			exit = true
		}
	}

	if err = ticker.stop(); err != nil {
		t.Errorf("ticker stop should not return an error, received %v", err)
	}

	if err = ticker.stop(); err == nil {
		t.Errorf("ticker stop should return an error")
	}
}

func TestNewTicker2(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(2)*time.Second)
	defer cancel()

	chTrigger := make(chan time.Time)
	ticker := NewTicker(ctx, EverySecond(), chTrigger)

	ticker.Stop()
	ticker.mux.RLock()
	if ticker.tickerCancel != nil { // Ticker should not be started yet
		t.Errorf("tickerCancel should be nil")
	}
	ticker.mux.RUnlock()

	ticker.Start()
	ticker.mux.RLock()
	if ticker.tickerCancel == nil {
		t.Errorf("tickerCancel should not be nil")
	}
	ticker.mux.RUnlock()

	var exit bool
	for {
		if exit {
			break
		}

		select {
		case <-ctx.Done():
			t.Errorf("ticker should have been triggered")
			exit = true
		case <-chTrigger:
			exit = true
		}
	}

	ticker.Stop()
	ticker.mux.RLock()
	if ticker.tickerCancel != nil {
		t.Errorf("tickerCancel should be nil")
	}
}

func TestNewTicker3(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	chTrigger := make(chan time.Time)
	ticker := NewTicker(ctx, EverySecond(), chTrigger)

	var err error
	if err = ticker.start(); err != nil {
		t.Errorf("ticker start should not return an error, received %v", err)
	}

	cancel()
	time.Sleep(time.Second)
	if err = ticker.stop(); err == nil {
		t.Errorf("ticker stop should return an error")
	}
}
