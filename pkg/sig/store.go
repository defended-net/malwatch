// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package sig

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/defended-net/malwatch/third_party/yr"
)

// Sigs represents sigs.
type Sigs struct {
	Rev   uint64
	Rules *yr.Rules
	refs  int64
	once  sync.Once
}

var curr atomic.Pointer[Sigs]

// New returns new sigs from given rev and rules.
func New(rules *yr.Rules, rev uint64) *Sigs {
	sigs := &Sigs{
		Rev:   rev,
		Rules: rules,
		refs:  1,
	}

	return sigs
}

// Acquire increments refs and returns sigs.
func Acquire() (*Sigs, error) {
	for {
		sigs := curr.Load()

		if sigs == nil {
			return nil, ErrYrcGet
		}

		refs := atomic.LoadInt64(&sigs.refs)
		if refs <= 0 {
			return nil, ErrYrcGet
		}

		if atomic.CompareAndSwapInt64(&sigs.refs, refs, refs+1) {
			return sigs, nil
		}

		// Cmp failed, retry.
		runtime.Gosched()
	}
}

// Set sets sigs from given yrc path and rev.
func Set(path string, rev uint64) error {
	rules, err := yr.LoadRules(path)
	if err != nil {
		return fmt.Errorf("%w, %v", ErrYrcLoad, err)
	}

	var (
		sigs = New(rules, rev)
		old  = curr.Swap(sigs)
	)

	if old != nil {
		old.Release()
	}

	return nil
}

// Release decrements refs. Use after Acquire.
func (sigs *Sigs) Release() {
	if atomic.AddInt64(&sigs.refs, -1) == 0 {
		sigs.destroy()
	}
}

// destroy destroys sigs.
func (sigs *Sigs) destroy() {
	sigs.once.Do(
		func() {
			if sigs.Rules != nil {
				sigs.Rules.Destroy()
			}
		},
	)
}
