// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package orm

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestGet(t *testing.T) {
	db, err := Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("db mock err %s", err)
	}

	input := struct{}{}

	if err := Put(db, "hits", t.Name(), input); err != nil {
		t.Fatalf("put err %s", err)
	}

	if _, got := Get(db, "hits", t.Name()); got != nil {
		t.Errorf("get err %v", got)
	}
}

func TestGetAll(t *testing.T) {
	db, err := Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("db mock err %s", err)
	}

	input := struct{}{}

	if err := Put(db, "hits", t.Name(), input); err != nil {
		t.Fatalf("put err %s", err)
	}

	if _, got := GetAll(db, "hits"); got != nil {
		t.Errorf("get all err %v", got)
	}
}

func TestGetErrs(t *testing.T) {
	want := ErrTxBktNotFound

	db, err := Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("db mock err %s", err)
	}

	if _, got := Get(db, t.Name(), t.Name()); !errors.Is(got, want) {
		t.Errorf("unexpected get err %v, want %v", got, want)
	}
}

func TestGetAllErrs(t *testing.T) {
	want := ErrBktIter

	db, err := Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("db mock err %s", err)
	}

	if _, got := GetAll(db, t.Name()); !errors.Is(got, want) {
		t.Errorf("unexpected get all err %v, want %v", got, want)
	}
}

func TestPutErrs(t *testing.T) {
	want := ErrTxBktNotFound

	db, err := Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("db mock err %s", err)
	}

	if got := Put(db, t.Name(), t.Name(), nil); !errors.Is(got, want) {
		t.Errorf("unexpected put err %v, want %v", got, want)
	}
}

func TestPutUnsupported(t *testing.T) {
	want := "unsupported type: chan string"

	db, err := Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("db mock err %s", err)
	}

	if got := Put(db, "hits", t.Name(), make(chan string)); !errors.Is(got, ErrMarshal) {
		t.Errorf("unexpected put err %v, want %v", got, want)
	}
}
