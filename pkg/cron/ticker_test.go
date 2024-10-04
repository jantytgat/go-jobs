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
	if err = ticker.Stop(); err == nil { // Ticker should not be started yet
		t.Errorf("ticker stop should return an error")
	}

	if err = ticker.Start(); err != nil {
		t.Errorf("ticker start should not return an error, received %v", err)
	}

	if err = ticker.Start(); err == nil { // try starting the goroutine again
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

	if err = ticker.Stop(); err != nil {
		t.Errorf("ticker stop should not return an error, received %v", err)
	}

	if err = ticker.Stop(); err == nil {
		t.Errorf("ticker stop should return an error")
	}
}

func TestNewTicker2(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	chTrigger := make(chan time.Time)
	ticker := NewTicker(ctx, EverySecond(), chTrigger)

	var err error
	if err = ticker.Start(); err != nil {
		t.Errorf("ticker start should not return an error, received %v", err)
	}

	cancel()
	time.Sleep(time.Second)
	if err = ticker.Stop(); err == nil {
		t.Errorf("ticker stop should return an error")
	}
}
