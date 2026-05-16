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
		input    = NewSpinner(interval, t.Name())
	)

	if input.msg != t.Name() {
		t.Errorf("unexpected spinner msg %v, want %v", input.msg, t.Name())
	}

	if input.interval != interval {
		t.Errorf("unexpected spinner interval %v, want %v", input.interval, interval)
	}
}

func TestStart(t *testing.T) {
	input := NewSpinner(100*time.Millisecond, t.Name())

	go func(spinner *Spinner) {
		spinner.Start()
	}(input)

	time.Sleep(500 * time.Millisecond)

	input.Stop()
}

func TestStartNil(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Fatalf("unexpected start panic")
		}
	}()

	var input *Spinner

	input.Start()
}

func TestStop(t *testing.T) {
	input := NewSpinner(100*time.Millisecond, t.Name())

	go func(spinner *Spinner) {
		spinner.Start()
	}(input)

	time.Sleep(500 * time.Millisecond)

	input.Stop()
}

func TestStopNil(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Fatalf("unexpected stop panic")
		}
	}()

	var input *Spinner

	input.Stop()
}
