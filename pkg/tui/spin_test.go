// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package tui

import (
	"testing"
	"time"
)

func TestNewSpinner(t *testing.T) {
	var (
		interval = 100 * time.Millisecond
		spinner  = NewSpinner(interval, t.Name())
	)

	if spinner.msg != t.Name() {
		t.Errorf("unexpected spinner msg: %v, want %v", spinner.msg, t.Name())
	}

	if spinner.interval != interval {
		t.Errorf("unexpected spinner interval: %v, want %v", spinner.interval, interval)
	}
}

func TestStart(t *testing.T) {
	spinner := NewSpinner(100*time.Millisecond, t.Name())

	go func(spinner *Spinner) {
		spinner.Start()
	}(spinner)

	time.Sleep(500 * time.Millisecond)

	spinner.Stop()
}

func TestStartNil(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Fatalf("unexpected panic during nil spinner start")
		}
	}()

	var input *Spinner

	input.Start()
}

func TestStop(t *testing.T) {
	spinner := NewSpinner(100*time.Millisecond, t.Name())

	go func(spinner *Spinner) {
		spinner.Start()
	}(spinner)

	time.Sleep(500 * time.Millisecond)

	spinner.Stop()
}

func TestStopNil(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Fatalf("unexpected panic during nil spinner stop")
		}
	}()

	var input *Spinner

	input.Stop()
}
