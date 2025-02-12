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
		t.Fatalf("db mock error: %s", err)
	}

	input := struct{}{}

	if err := Put(db, "hits", t.Name(), input); err != nil {
		t.Fatalf("put error: %s", err)
	}

	if _, err = Get(db, "hits", t.Name()); err != nil {
		t.Errorf("get error: %v", err)
	}
}

func TestGetAll(t *testing.T) {
	db, err := Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("db mock error: %s", err)
	}

	input := struct{}{}

	if err := Put(db, "hits", t.Name(), input); err != nil {
		t.Fatalf("put error: %s", err)
	}

	if _, err = GetAll(db, "hits"); err != nil {
		t.Errorf("get all error: %v", err)
	}
}

func TestGetErrs(t *testing.T) {
	db, err := Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("db mock error: %s", err)
	}

	if _, err = Get(db, t.Name(), t.Name()); !errors.Is(err, ErrTxBktNotFound) {
		t.Errorf("unexpected get error: %v, want %v", err, ErrTxBktNotFound)
	}
}

func TestGetAllErrs(t *testing.T) {
	db, err := Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("db mock error: %s", err)
	}

	if _, err = GetAll(db, t.Name()); !errors.Is(err, ErrBktIter) {
		t.Errorf("unexpected get error: %v, want %v", err, ErrBktIter)
	}
}

func TestPutErrs(t *testing.T) {
	db, err := Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("db mock error: %s", err)
	}

	if err = Put(db, t.Name(), t.Name(), nil); !errors.Is(err, ErrTxBktNotFound) {
		t.Errorf("unexpected get error: %v, want %v", err, ErrTxBktNotFound)
	}
}

func TestPutUnsupported(t *testing.T) {
	db, err := Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("db mock error: %s", err)
	}

	want := "unsupported type: chan string"

	if err = Put(db, "hits", t.Name(), make(chan string)); !errors.Is(err, ErrMarshal) {
		t.Errorf("unexpected get error: %v, want %v", err, want)
	}
}
